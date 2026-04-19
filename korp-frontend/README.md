# Korp Frontend

Frontend Angular da aplicação de estoque e faturamento.

## Ambientes

Arquivos de configuração:

- `src/environments/environment.ts` (dev)
- `src/environments/environment.homolog.ts` (homolog)
- `src/environments/environment.prod.ts` (produção)

Build por ambiente:

```bash
npm run build -- --configuration development
npm run build -- --configuration homolog
npm run build -- --configuration production
```

## Execução local

```bash
npm install
npm run start
```

Aplicação disponível em `http://localhost:4200`.

## Testes unitários

```bash
npm run test -- --watch=false --progress=false
```

## Testes E2E (Playwright)

Instalação do browser (primeira vez):

```bash
npm run test:e2e:install
```

Execução da suíte:

```bash
npm run test:e2e
```

Modo UI (debug local):

```bash
npm run test:e2e:ui
```
