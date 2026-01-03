#!/usr/bin/env python3
import os
import re
import base64
import argparse
from datetime import datetime, timezone
from typing import Optional, Tuple, List, Dict, Any

from bs4 import BeautifulSoup
from markdownify import markdownify as md

from googleapiclient.discovery import build
from googleapiclient.errors import HttpError
from google.auth.transport.requests import Request
from google.oauth2.credentials import Credentials
from google_auth_oauthlib.flow import InstalledAppFlow

SCOPES = ["https://www.googleapis.com/auth/gmail.readonly"]


def ensure_dir(path: str) -> None:
    os.makedirs(path, exist_ok=True)


def sanitize_filename(name: str, max_len: int = 140) -> str:
    name = (name or "").strip()
    # remove chars inválidos pra filename
    name = re.sub(r"[\\/:*?\"<>|]+", "-", name)
    name = re.sub(r"\s+", " ", name)
    name = name[:max_len].strip()
    return name if name else "sem-titulo"


def yaml_safe(s: str) -> str:
    """Evita quebrar YAML com aspas duplas; troca por aspas simples."""
    return (s or "").replace('"', "'").strip()


def load_gmail_service(credentials_path: str, token_path: str):
    creds = None

    if os.path.exists(token_path):
        creds = Credentials.from_authorized_user_file(token_path, SCOPES)

    if not creds or not creds.valid:
        if creds and creds.expired and creds.refresh_token:
            creds.refresh(Request())
        else:
            flow = InstalledAppFlow.from_client_secrets_file(credentials_path, SCOPES)
            creds = flow.run_local_server(port=0)

        with open(token_path, "w", encoding="utf-8") as f:
            f.write(creds.to_json())

    return build("gmail", "v1", credentials=creds)


def decode_b64url(data: str) -> str:
    """Gmail usa base64url (urlsafe)."""
    if not data:
        return ""
    missing_padding = len(data) % 4
    if missing_padding:
        data += "=" * (4 - missing_padding)
    return base64.urlsafe_b64decode(data.encode("utf-8")).decode("utf-8", errors="replace")


def find_header(headers: List[Dict[str, str]], name: str) -> str:
    name_lower = name.lower()
    for h in headers or []:
        if (h.get("name") or "").lower() == name_lower:
            return h.get("value") or ""
    return ""


def extract_best_body(payload: Dict[str, Any]) -> Tuple[Optional[str], Optional[str]]:
    """
    Retorna (html, text) priorizando:
    - text/html
    - text/plain
    Percorre recursivamente multiparts.
    """
    if not payload:
        return None, None

    mime = payload.get("mimeType")
    body = payload.get("body") or {}
    data = body.get("data")

    if mime == "text/html" and data:
        return decode_b64url(data), None

    if mime == "text/plain" and data:
        return None, decode_b64url(data)

    parts = payload.get("parts") or []
    html_candidate = None
    text_candidate = None

    for p in parts:
        h, t = extract_best_body(p)
        if h and not html_candidate:
            html_candidate = h
        if t and not text_candidate:
            text_candidate = t

    return html_candidate, text_candidate


def html_to_markdown(html: str) -> str:
    soup = BeautifulSoup(html or "", "html.parser")

    # remove scripts/styles
    for tag in soup(["script", "style", "noscript"]):
        tag.decompose()

    cleaned_html = str(soup)
    return md(cleaned_html, heading_style="ATX")


def parse_internal_date(ms: Optional[str]) -> str:
    if not ms:
        return ""
    try:
        dt = datetime.fromtimestamp(int(ms) / 1000, tz=timezone.utc).astimezone()
        return dt.strftime("%Y-%m-%d %H:%M:%S %z")
    except Exception:
        return ""


def gmail_list_all(service, query: str, max_results: int = 5000):
    """
    Lista mensagens que batem no query, paginando até max_results.
    """
    user_id = "me"
    messages = []
    page_token = None

    while True:
        remaining = max_results - len(messages)
        if remaining <= 0:
            break

        resp = (
            service.users()
            .messages()
            .list(
                userId=user_id,
                q=query,
                pageToken=page_token,
                maxResults=min(500, remaining),
            )
            .execute()
        )

        batch = resp.get("messages") or []
        messages.extend(batch)

        page_token = resp.get("nextPageToken")
        if not page_token:
            break

    return messages


def gmail_get_message(service, msg_id: str):
    return service.users().messages().get(userId="me", id=msg_id, format="full").execute()


def build_markdown(
    subject: str,
    from_: str,
    to: str,
    date_str: str,
    message_id: str,
    content_md: str,
) -> str:
    safe_subject = yaml_safe(subject)
    safe_from = yaml_safe(from_)
    safe_to = yaml_safe(to)
    safe_message_id = yaml_safe(message_id)

    fm = [
        "---",
        f'title: "{safe_subject}"',
        f'from: "{safe_from}"',
        f'to: "{safe_to}"',
        f'date: "{yaml_safe(date_str)}"' if date_str else 'date: ""',
        f'message_id: "{safe_message_id}"',
        "source: gmail",
        "---",
        "",
        f"# {safe_subject}",
        "",
    ]

    return "\n".join(fm) + (content_md or "").strip() + "\n"


def main():
    parser = argparse.ArgumentParser(
        description="Baixar emails do Gmail como Markdown (1 email = 1 capítulo)."
    )
    parser.add_argument(
        "--query",
        required=True,
        help='Busca do Gmail. Ex: \'subject:"Capítulo" from:me newer_than:1y\'',
    )
    parser.add_argument("--out", default="chapters_md", help="Pasta de saída")
    parser.add_argument("--max", type=int, default=500, help="Máximo de emails para baixar")
    parser.add_argument("--credentials", default="credentials.json", help="Arquivo credentials.json do OAuth")
    parser.add_argument("--token", default="token.json", help="Arquivo token.json gerado após login")
    parser.add_argument("--prefix", default="", help="Prefixo no nome do arquivo (opcional)")
    args = parser.parse_args()

    if not os.path.exists(args.credentials):
        raise SystemExit(
            f'Não achei "{args.credentials}". Coloque o credentials.json (OAuth Desktop) na mesma pasta do script.'
        )

    ensure_dir(args.out)

    service = load_gmail_service(args.credentials, args.token)

    print(f"🔎 Buscando emails com query: {args.query}")
    msg_refs = gmail_list_all(service, args.query, max_results=args.max)

    if not msg_refs:
        print("Nenhum email encontrado.")
        return

    print(f"📩 Encontrados: {len(msg_refs)} emails. Lendo conteúdo...")

    full_msgs = []
    for m in msg_refs:
        try:
            full_msgs.append(gmail_get_message(service, m["id"]))
        except HttpError as e:
            print(f"⚠️ Erro ao ler msg {m.get('id')}: {e}")

    # ordena por data interna
    full_msgs.sort(key=lambda x: int(x.get("internalDate", "0")))

    print(f"💾 Salvando em: {args.out}")

    for idx, msg in enumerate(full_msgs, start=1):
        payload = msg.get("payload") or {}
        headers = payload.get("headers") or []

        subject = find_header(headers, "Subject") or f"Sem assunto {idx}"
        from_ = find_header(headers, "From")
        to = find_header(headers, "To")
        date_str = find_header(headers, "Date") or parse_internal_date(msg.get("internalDate"))
        message_id = find_header(headers, "Message-Id")

        html, text = extract_best_body(payload)

        if html:
            content_md = html_to_markdown(html)
        elif text:
            content_md = text
        else:
            content_md = ""

        chapter_md = build_markdown(subject, from_, to, date_str, message_id, content_md)

        safe_name = sanitize_filename(subject)
        filename = f"{idx:03d} - {safe_name}.md"
        if args.prefix:
            filename = f"{args.prefix}{filename}"

        path = os.path.join(args.out, filename)

        with open(path, "w", encoding="utf-8") as f:
            f.write(chapter_md)

        print(f"✅ [{idx:03d}] {subject}")

    print("🎉 Finalizado! Seus capítulos estão em Markdown.")


if __name__ == "__main__":
    main()
