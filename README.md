# Teste Prático Korp - Documentação Técnica

## Tópico da documentação

Documentação técnica completa do projeto, com foco em execução local, regras de negócio, validação por rotas, testes automatizados, observabilidade e publicação final.

## 1) Visão geral

Este repositório contém um sistema completo de estoque e faturamento, dividido em 3 aplicações:

- `estoque`: microserviço Go responsável pelo cadastro de produtos e baixa de saldo.
- `faturamento`: microserviço Go responsável pelo ciclo de vida de notas fiscais.
- `korp-frontend`: aplicação Angular que consome os dois serviços.

Fluxo principal:

1. Cadastrar produto no serviço de estoque.
2. Criar nota fiscal no serviço de faturamento.
3. Imprimir nota fiscal.
4. Ao imprimir, o faturamento chama o estoque para baixar saldo dos produtos.

## 2) Arquitetura e portas

- Frontend Angular: `http://localhost:4200`
- API de estoque: `http://localhost:8080`
- API de faturamento: `http://localhost:8081`
- Banco de dados: PostgreSQL (via `DB_URL` em cada serviço)

Integração entre serviços:

- O faturamento chama o estoque em `POST /produtos/:id/baixa`.
- URL base da integração é configurada por `ESTOQUE_SERVICE_URL`.

## 3) Estrutura de pastas

```text
Korp_Teste_Eduardo/
 estoque/
  controllers/
  db/
  middlewares/
  models/
  main.go
 faturamento/
  controllers/
  db/
  middlewares/
  models/
  main.go
 korp-frontend/
  src/
   app/
   environments/
  e2e/
  angular.json
  package.json
 scripts/
  start-air-estoque.ps1
  start-air-faturamento.ps1
 docs/
  OPERACAO_RELEASE.md
```

## 4) Requisitos

- Go `1.26.2`
- Node.js `18+` (recomendado `20+`)
- npm `11+`
- PostgreSQL em execução
- Git

Opcional:

- Air (hot reload para Go)

## 5) Configuração de ambiente

## 5.1 Backend de estoque (`estoque/.env`)

```env
DB_URL=postgres://usuario:senha@localhost:5432/estoque?sslmode=disable
ALLOW_ORIGINS=http://localhost:4200
```

## 5.2 Backend de faturamento (`faturamento/.env`)

```env
DB_URL=postgres://usuario:senha@localhost:5432/faturamento?sslmode=disable
ALLOW_ORIGINS=http://localhost:4200
ESTOQUE_SERVICE_URL=http://localhost:8080
```

## 5.3 Frontend Angular

Os arquivos de ambiente ficam em `korp-frontend/src/environments`:

- `environment.ts` (desenvolvimento)
- `environment.homolog.ts` (homologação)
- `environment.prod.ts` (produção)

Antes da publicação final, ajuste as URLs reais de homologação e produção nesses arquivos.

## 6) Como executar localmente

## 6.1 Subir o estoque

```powershell
cd .\estoque
go mod tidy
go run main.go
```

## 6.2 Subir o faturamento

```powershell
cd .\faturamento
go mod tidy
go run main.go
```

## 6.3 Subir o frontend

```powershell
cd .\korp-frontend
npm install
ng serve
```

Depois disso, acesse `http://localhost:4200`.

## 6.4 Hot reload com Air (Windows)

Instalação (uma vez):

```powershell
go install github.com/air-verse/air@latest
```

Executar estoque:

```powershell
.\scripts\start-air-estoque.ps1
```

Executar faturamento:

```powershell
.\scripts\start-air-faturamento.ps1
```

## 7) APIs e regras de negócio

## 7.1 Estoque (`:8080`)

### `POST /produtos`

Cria produto.

Regras:

- `codigo` obrigatório e maior que zero
- `descricao` obrigatória
- `saldo` não pode ser negativo
- `preco` não pode ser negativo
- bloqueia `codigo` duplicado

### `GET /produtos`

Lista produtos ordenados por código crescente.

### `PUT /produtos/:id`

Atualiza descrição, saldo e preço.

Regra:

- código é imutável

### `DELETE /produtos/:id`

Remove produto pelo código.

### `POST /produtos/:id/baixa`

Baixa saldo de estoque.

Regra:

- quantidade deve ser maior que zero
- não permite saldo insuficiente

## 7.2 Faturamento (`:8081`)

### `POST /notas-fiscais`

Cria nota fiscal com cliente e itens.

Regras:

- cliente obrigatório
- ao menos um item
- item com `produto_id` válido
- item com `quantidade` maior que zero
- nota inicia com status aberta

### `GET /notas-fiscais`

Lista notas fiscais com itens.

### `POST /notas-fiscais/:id/imprimir`

Fecha a nota e baixa estoque item a item.

Regras:

- Apenas nota aberta pode ser impressa
- Em falha de integração com estoque, nota permanece aberta

### `POST /notas-fiscais/:id/cancelar`

Cancela nota aberta.

Regra:

- Apenas nota aberta pode ser cancelada

### `DELETE /notas-fiscais/:id`

Exclui nota fechada e seus itens em transação.

Regra:

- Nota aberta não pode ser excluída

## 8) Roteiro de validação final por rotas

## 8.1 Produtos: cadastro, edição e exclusão

Rotas:

- Frontend: `http://localhost:4200/produtos`
- `POST http://localhost:8080/produtos`
- `GET http://localhost:8080/produtos`
- `PUT http://localhost:8080/produtos/:id`
- `DELETE http://localhost:8080/produtos/:id`

Passo a passo:

1. Cadastrar produto válido.
2. Confirmar exibição na lista.
3. Editar descrição/saldo/preço.
4. Excluir no modal de confirmação.
5. Validar mensagens de sucesso e erro.

## 8.2 Notas fiscais: criação e impressão

Rotas:

- Frontend: `http://localhost:4200/notas-fiscais`
- `POST http://localhost:8081/notas-fiscais`
- `GET http://localhost:8081/notas-fiscais`
- `POST http://localhost:8081/notas-fiscais/:id/imprimir`

Passo a passo:

1. Criar nota com cliente e itens.
2. Confirmar status inicial aberta.
3. Imprimir a nota.
4. Confirmar status fechada.
5. Confirmar baixa no estoque.

## 8.3 Notas fiscais: cancelamento e exclusão por status

Rotas:

- `POST http://localhost:8081/notas-fiscais/:id/cancelar`
- `DELETE http://localhost:8081/notas-fiscais/:id`

Passo a passo:

1. Cancelar uma nota aberta.
2. Tentar cancelar nota fechada.
3. Excluir nota fechada.
4. Tentar excluir nota aberta.

## 8.4 Falha entre microsserviços

Passo a passo:

1. Parar o serviço de estoque.
2. Tentar imprimir nota aberta.
3. Validar mensagem amigável de erro.
4. Confirmar que a nota continua aberta.

## 9) Testes automatizados

## 9.1 Frontend unitário (Vitest)

```powershell
cd .\korp-frontend
npm run test -- --watch=false --progress=false
```

## 9.2 Frontend E2E (Playwright)

Instalação do navegador (primeira vez):

```powershell
cd .\korp-frontend
npm run test:e2e:install
```

Execução:

```powershell
cd .\korp-frontend
npm run test:e2e
```

Modo interface gráfica:

```powershell
cd .\korp-frontend
npm run test:e2e:ui
```

## 9.3 Backend (Go)

```powershell
cd .\estoque
go test ./...

cd ..\faturamento
go test ./...
```

## 9.4 Build de validação do frontend

```powershell
cd .\korp-frontend
npm run build -- --configuration development
```

## 10) Observabilidade

O projeto possui rastreabilidade por requisição com `request_id`.

Comportamento:

- Cada resposta inclui header `X-Request-Id`.
- Respostas de erro incluem `request_id` no JSON.
- Logs registram:
  - serviço
  - método
  - rota
  - status
  - duração em ms
  - IP
  - código/mensagem de erro (quando houver)

## 11) Detalhamento técnico solicitado para avaliação

### 11.1 Ciclos de vida do Angular utilizados

- Foi utilizado o ciclo de vida `OnInit` com `ngOnInit()` nos componentes de tela:
  - `Produtos`
  - `NotasFiscais`
- O objetivo do `ngOnInit()` nesses componentes é carregar os dados iniciais (listas de produtos e notas) assim que a tela é inicializada.

### 11.2 Uso de RxJS

Sim. O uso de RxJS acontece por meio dos `Observable`s retornados pelo `HttpClient` do Angular.

- As chamadas HTTP usam `subscribe({ next, error, complete })` para:
  - atualizar estado da tela em sucesso;
  - tratar erro com mensagem amigável;
  - finalizar estados de carregamento/processamento.
- Neste projeto, não foram aplicados operadores RxJS mais avançados (como `map`, `switchMap`, `mergeMap`), pois o fluxo de API é direto.

### 11.3 Outras bibliotecas utilizadas e finalidade

Frontend:

- `@angular/router`: navegação entre rotas da aplicação.
- `@angular/forms`: binding e validação de formulários nos componentes.
- `@angular/common` (incluindo `CurrencyPipe`): utilitários comuns e formatação monetária.
- `@playwright/test`: testes E2E dos fluxos críticos.
- `vitest` + `jsdom`: testes unitários no frontend.
- `prettier`: padronização de formatação do código.

Backend (Go):

- `github.com/gin-gonic/gin`: framework HTTP para APIs REST.
- `github.com/gin-contrib/cors`: middleware CORS.
- `gorm.io/gorm`: ORM para acesso ao banco de dados.
- `gorm.io/driver/postgres`: driver PostgreSQL para o GORM.
- `github.com/joho/godotenv`: carga de variáveis de ambiente locais via arquivo `.env`.
- `github.com/glebarez/sqlite` e `github.com/glebarez/go-sqlite`: suporte a banco em memória para testes automatizados.

### 11.4 Bibliotecas de componentes visuais

- Não foi utilizada biblioteca de componentes visuais pronta (como Angular Material, PrimeNG ou Bootstrap).
- A interface foi implementada com templates Angular + CSS próprio.
- Para exibição de valores monetários, foi utilizado `CurrencyPipe` do Angular (`@angular/common`).

### 11.5 Gerenciamento de dependências no Golang

- Foi utilizado Go Modules, com controle por `go.mod` e `go.sum` em cada serviço (`estoque` e `faturamento`).
- O comando `go mod tidy` foi usado para sincronizar e limpar dependências conforme os imports do código.
- Dependências de execução e de teste ficam versionadas nos arquivos de módulo.

### 11.6 Frameworks utilizados no Golang

- Golang:
  - `Gin` (camada web/API REST).
  - `GORM` (persistência e transações com banco relacional).
- C#:
  - Não aplicável neste projeto (não há implementação em C#).

### 11.7 Tratamento de erros e exceções no backend

- Validações de entrada retornam erros de cliente (`400`) com código de erro semântico.
- Regras de negócio (ex.: conflito de estado, código duplicado, estoque insuficiente) retornam `409`.
- Recursos inexistentes retornam `404`.
- Falhas internas (banco/integração) retornam `500`.
- As respostas de erro seguem padrão JSON com campos como:
  - `codigo`
  - `erro`
  - `detalhes` (quando aplicável)
  - `request_id` (para rastreio)
- Foi utilizado `gin.Recovery()` para captura de pânicos na camada HTTP.
- Fluxos críticos usam transação no GORM para manter consistência em caso de falha.

## 12) Publicação

Recomendação prática:

1. Publicar backends (`estoque` e `faturamento`) em um provedor de backend.
2. Ajustar URLs reais em `environment.homolog.ts` e `environment.prod.ts`.
3. Publicar frontend (por exemplo, Vercel).
4. Configurar `ALLOW_ORIGINS` dos backends com o domínio publicado do frontend.

## 13) Problemas comuns e solução rápida

### Erro de CORS

- Verifique `ALLOW_ORIGINS` nos dois backends.
- Inclua exatamente a origem do frontend.

### Falha ao imprimir nota

- Verifique se `ESTOQUE_SERVICE_URL` aponta para o estoque correto.
- Confira logs e `request_id` para rastrear a falha.

### Falha de conexão com banco

- Verifique `DB_URL` e se o PostgreSQL está ativo.

### Comandos npm com parâmetros

- Se houver comportamento estranho no script, rode com `npx ng ...` como alternativa.

## 14) Documentos complementares

- Guia de operação e rollback: `docs/OPERACAO_RELEASE.md`
