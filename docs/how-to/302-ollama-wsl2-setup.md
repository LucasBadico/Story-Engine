# Ollama Setup (Windows + WSL2)

> Configure Ollama running on Windows to serve embeddings to the LLM Gateway running in WSL2.

---

## Overview

The LLM Gateway uses Ollama for embedding generation (`nomic-embed-text`). Ollama runs on Windows (with GPU access), while the gateway runs in WSL2. WSL2 reaches the Windows host via an internal network IP.

---

## 1. Install Ollama on Windows

Download from https://ollama.com/download/windows and install.

### Pull the embedding model

Open PowerShell:

```powershell
ollama pull nomic-embed-text
```

Verify:

```powershell
ollama list
# Should show nomic-embed-text
```

---

## 2. Configure Ollama to Listen on All Interfaces

By default Ollama only listens on `127.0.0.1` (Windows localhost), which WSL2 cannot reach.

### Option A: Environment variable (per session)

```powershell
$env:OLLAMA_HOST="0.0.0.0"
ollama serve
```

### Option B: System environment variable (persistent)

1. Open **Settings > System > About > Advanced system settings > Environment Variables**
2. Under **User variables**, click **New**
3. Variable name: `OLLAMA_HOST`
4. Variable value: `0.0.0.0`
5. Click OK, restart Ollama

### Verify Ollama is accessible from WSL2

In WSL2:

```bash
curl http://win-3090:11434/api/tags
# Should return JSON with model list
```

If using Tailscale, the hostname `win-3090` resolves via MagicDNS and is stable across reboots. If not using Tailscale, use the WSL2 gateway IP instead: `ip route show default | awk '{print $3}'` (this IP can change after reboots).

---

## 3. Configure the LLM Gateway

Update `llm-gateway-service/.env`:

```
EMBEDDING_PROVIDER=ollama
EMBEDDING_MODEL=nomic-embed-text
EMBEDDING_BASE_URL=http://win-3090:11434
```

After changing `.env`, restart the gateway worker:

```bash
cd llm-gateway-service && make worker-blocks-llm
```

---

## Known Issues

### Using Tailscale (recommended)

With Tailscale installed on both Windows and WSL2, use the MagicDNS hostname `win-3090`. This is stable and doesn't change across reboots.

### Without Tailscale: Windows host IP changes after WSL restart

The IP at `ip route show default` can change after `wsl --shutdown` or a Windows reboot. If the worker fails with connection errors, re-check the IP and update `.env`.

Quick fix:

```bash
WINDOWS_IP=$(ip route show default | awk '{print $3}')
sed -i "s|EMBEDDING_BASE_URL=.*|EMBEDDING_BASE_URL=http://$WINDOWS_IP:11434|" llm-gateway-service/.env
```

### Windows Firewall blocks WSL2

If `curl` to the Windows IP times out, Windows Firewall may be blocking inbound connections from WSL2. Add a rule:

```powershell
# Run as Administrator in PowerShell
New-NetFirewallRule -DisplayName "Ollama WSL2" -Direction Inbound -LocalPort 11434 -Protocol TCP -Action Allow
```

### Ollama not found after install

If `ollama` isn't in PATH, the default install location is `%LOCALAPPDATA%\Programs\Ollama`. Add it to your PATH or run it from there.

---

## Quick Reference

| What | Command |
|------|---------|
| Pull model | `ollama pull nomic-embed-text` (Windows) |
| Start Ollama (all interfaces) | `$env:OLLAMA_HOST="0.0.0.0"; ollama serve` (Windows) |
| Test from WSL2 (Tailscale) | `curl http://win-3090:11434/api/tags` |
| Get Windows IP (no Tailscale) | `ip route show default \| awk '{print $3}'` |
| Restart gateway worker | `cd llm-gateway-service && make worker-blocks-llm` |
