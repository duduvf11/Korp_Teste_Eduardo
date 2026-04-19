# Korp_Teste_Eduardo

## Hot Reload no Go com Air (Windows)

O Air ja foi configurado nos dois microsservicos:

- `estoque/.air.toml`
- `faturamento/.air.toml`

Tambem foram criados scripts para iniciar o Air com ajuste automatico de PATH no Windows:

- `scripts/start-air-estoque.ps1`
- `scripts/start-air-faturamento.ps1`

### 1) Instalar Air (uma vez)

```powershell
go install github.com/air-verse/air@latest
```

### 2) Rodar o microsservico de estoque com hot reload

```powershell
.
\scripts\start-air-estoque.ps1
```

### 3) Rodar o microsservico de faturamento com hot reload

```powershell
.
\scripts\start-air-faturamento.ps1
```

### 4) Como funciona

Sempre que voce salvar um arquivo `.go`, o Air recompila e reinicia o servico automaticamente.

### 5) Se PowerShell classico nao estiver no PATH

Se aparecer erro `exec: "powershell": executable file not found in %PATH%`, os scripts acima ja corrigem isso automaticamente na sessao atual.

### 6) Parar o processo

Use `Ctrl + C` no terminal de cada microsservico.
