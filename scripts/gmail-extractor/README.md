# Gmail Extractor

Script para baixar emails do Gmail e convertê-los automaticamente para arquivos Markdown (um email por arquivo) com frontmatter pronto para ingestão no Story Engine.

## Requisitos

- Python 3.9+ com `pip`
- Uso obrigatório de ambiente virtual dedicado (`.venv`) para instalar dependências
- Acesso a uma conta do Gmail com os emails que você quer extrair

## Preparando o ambiente virtual (obrigatório)

> Não consigo executar comandos aqui; rode no seu terminal e me avise o resultado se algo falhar.

1. Na raiz do repositório, crie a venv:
   ```bash
   python3 -m venv .venv
   ```
2. Ative a venv sempre antes de qualquer comando relacionado ao extractor:
   ```bash
   source .venv/bin/activate
   ```
3. *Enquanto a venv estiver ativa*, instale as dependências:
   ```bash
   pip install -U google-api-python-client google-auth google-auth-oauthlib bs4 markdownify
   ```
4. Toda nova sessão de terminal exige o passo 2 novamente (`source .venv/bin/activate`). Se precisar atualizar dependências, mantenha a venv ativa e repita o passo 3.

## Preparando as credenciais do Gmail API

1. Acesse [console.cloud.google.com](https://console.cloud.google.com/) e crie um projeto (ou use um existente).
2. Habilite a **Gmail API** no projeto.
3. Crie um **OAuth Client ID** do tipo *Desktop App* e faça o download do `credentials.json`.
4. Salve o `credentials.json` na pasta `scripts/gmail-extractor/` (mesmo local do `main.py`).
5. O script gera automaticamente um `token.json` na primeira execução (fluxo OAuth interativo). Ele guarda o refresh token para as próximas execuções.

## Estrutura de saída

- Os arquivos Markdown são criados dentro da pasta definida por `--out` (padrão `chapters_md/`).
- Cada arquivo recebe nome `NNN - <assunto>.md`, respeitando a ordem cronológica.
- O conteúdo inclui frontmatter YAML com remetente, destinatário, data, `message_id` e fonte (`gmail`), seguido da versão Markdown do corpo do email (HTML convertido quando disponível).

## Executando

> ⚠️ Rodar sem a venv ativa não é suportado. Garanta que `source .venv/bin/activate` foi executado na sessão atual.

Na raiz do repositório (ou dentro de `scripts/gmail-extractor/`), com a venv ativa, execute:

```bash
python scripts/gmail-extractor/main.py \
  --query 'subject:"Capítulo" from:me newer_than:1y' \
  --out chapters_md \
  --max 200 \
  --credentials scripts/gmail-extractor/credentials.json \
  --token scripts/gmail-extractor/token.json \
  --prefix ''
```

Parâmetros principais:

- `--query` **(obrigatório)**: string de busca padrão do Gmail (mesma da caixa de entrada). Exemplos:
  - `'from:eu@example.com subject:"Capítulo" newer_than:6m'`
  - `'label:Rascunhos after:2023/01/01 before:2024/01/01'`
- `--out`: pasta onde os .md serão salvos (criada automaticamente se não existir).
- `--max`: número máximo de emails a baixar (default 500).
- `--credentials`: caminho do `credentials.json`.
- `--token`: caminho para gravar o `token.json` (reutilizado depois).
- `--prefix`: string opcional adicionada ao início do nome de cada arquivo (ex: `cap-`).

## Fluxo típico

1. Configurar dependências e credenciais (passos anteriores).
2. Definir o filtro `--query` que representa a sua coleção de capítulos/emails.
3. Rodar o script e autorizar o acesso no navegador na primeira execução.
4. Revisar os `.md` gerados em `chapters_md/` (ou pasta definida) e importar no Story Engine.

## Dicas e resolução de problemas

- **Nada é encontrado**: teste o mesmo `--query` direto no Gmail web e ajuste até retornar os emails esperados.
- **Erro de credencial**: confirme se está usando OAuth do tipo *Desktop App* e se o arquivo foi salvo no caminho indicado.
- **Formato estranho no markdown**: o script prioriza corpo HTML. Se o email vier apenas em texto plano a conversão é direta; revise as mensagens originais.
- **Quota/limites**: a Gmail API tem limites. Se receber `HttpError 429`, reduza `--max` ou tente novamente depois.

Precisando de variações (ex: exportar anexos, mudar o frontmatter), me avise que orientamos como adaptar o `main.py`.

