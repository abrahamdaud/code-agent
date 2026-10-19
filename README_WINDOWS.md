# Windows instructions for running CodeAgent

This document explains how to run CodeAgent on Windows / PowerShell, how to enable a terminal environment that works well with TUI libraries (ANSI/VT support), and how to run the non-TUI fallback CLI when a modern terminal is not available.

Prerequisites
- Go 1.22+ installed and on PATH
- Git
- Prefer using Windows Terminal or another modern terminal emulator (recommended). Older PowerShell/ConHost may not support full ANSI/VT features.

1) Build the project
Open PowerShell and from repo root:

```powershell
git fetch origin
git checkout cross-platform-compat
go mod tidy
go build -o codeagent.exe ./...
```

2) Use the supplied PowerShell launcher
A helper script is provided at `scripts/run-windows.ps1`. The launcher sets UTF-8 and a TERM hint and then runs a command.

Examples:
- Run provider CLI (binary):
  ```powershell
  $env:HUGGINGFACE_API_KEY = 'hf_...'
  .\scripts\run-windows.ps1 -Cmd .\codeagent.exe -Args "provider-run -provider huggingface -model distilgpt2 -prompt \"Write a short Go hello world\""
  ```

- Run provider CLI via `go run` (no build):
  ```powershell
  $env:HUGGINGFACE_API_KEY = 'hf_...'
  .\scripts\run-windows.ps1 -Cmd "go" -Args "run .\cmd\provider_run -provider huggingface -model distilgpt2 -prompt \"Hello\""
  ```

3) If you prefer to run the non-TUI fallback CLI (plain mode)
Sometimes older terminals cannot render TUI correctly. A plain non-interactive CLI is provided in `cmd/no_tui` which reads a prompt and prints the provider response.

- Build and run:
  ```powershell
  go run ./cmd/no_tui -provider huggingface -model distilgpt2 -prompt "Write a short Go hello world"
  ```

- Or run interactively (prompt from stdin):
  ```powershell
  go run ./cmd/no_tui -provider huggingface -model distilgpt2
  # you will be prompted to enter the prompt text
  ```

4) Tips to improve terminal compatibility
- Enable UTF-8:
  ```powershell
  chcp 65001
  ```
- Use Windows Terminal and ensure it is updated.
- If you still see incorrect colors or broken UI, run the plain CLI from `cmd/no_tui`.

5) Troubleshooting
- If you see errors about API keys, ensure the appropriate env var is set (e.g. `HUGGINGFACE_API_KEY`, `OPENROUTER_API_KEY`, `OPENAI_API_KEY`).
- If build fails due to sqlite/CGO issues, either install `libsqlite3` dev tools (MSYS2/GCC) or ask to switch to the pure-Go sqlite driver (I can provide a patch).

6) Security
- Never commit API keys to source control. Use environment variables or the encrypted credential store.

If you'd like, I can also add a short README section to the project's main README to reference this file.