# Operação e Release - Teste Prático Korp

## Tópico da documentação

Guia operacional de hardening com estratégia de rollback, checklist de release e padronização de mensagens para operação.

## Estratégia de rollback

### 1) Rollback do frontend

Objetivo:

- voltar rapidamente para a última versão estável quando houver falha de navegação, integração ou regressão visual.

Passo a passo:

1. Identificar o último commit estável publicado do frontend.
2. Gerar novo build a partir desse commit.
3. Publicar o artefato anterior no ambiente afetado.
4. Validar as rotas:
   - `http://localhost:4200/produtos`
   - `http://localhost:4200/notas-fiscais`
5. Confirmar logs sem crescimento anormal de erro.

### 2) Rollback dos microsserviços

Objetivo:

- restaurar rapidamente o comportamento anterior dos serviços de estoque e faturamento.

Passo a passo:

1. Restaurar imagem/binário da última versão estável de `estoque` e `faturamento`.
2. Subir os serviços com configuração conhecida de ambiente.
3. Validar rotas críticas:
   - `POST /produtos`
   - `POST /notas-fiscais`
   - `POST /notas-fiscais/:id/imprimir`
4. Conferir `X-Request-Id` e logs de requisição para rastreio pós-rollback.

## Checklist final de release

### Pré-release

1. Executar testes unitários frontend: `npm run test -- --watch=false --progress=false`.
2. Executar build frontend: `npm run build -- --configuration development`.
3. Executar suíte E2E: `npm run test:e2e`.
4. Executar backend estoque: `cd estoque && go test ./...`.
5. Executar backend faturamento: `cd faturamento && go test ./...`.
6. Revisar URLs dos ambientes em:
   - `korp-frontend/src/environments/environment.homolog.ts`
   - `korp-frontend/src/environments/environment.prod.ts`

### Go-live

1. Publicar backend e frontend.
2. Testar rotas de produtos e notas em ambiente alvo.
3. Validar cenário de impressão com atualização de estoque.
4. Validar cenário de falha de impressão (mensagem amigável e nota aberta).

### Pós-release

1. Monitorar taxa de erro HTTP 4xx/5xx.
2. Verificar logs com `request_id` para correlação de incidentes.
3. Registrar incidentes e plano de ação corretiva.

## Padronização de mensagens para operação

### Backend de estoque

| Código | Status | Mensagem padrão | Ação operacional |
| --- | --- | --- | --- |
| `FORMATO_INVALIDO` | 400 | Dados inválidos para cadastro/baixa | Validar payload enviado pelo cliente |
| `CODIGO_INVALIDO` | 400 | Código do produto inválido | Corrigir código e reenviar |
| `CODIGO_DUPLICADO` | 409 | Já existe produto com esse código | Reutilizar cadastro existente ou informar outro código |
| `SALDO_INSUFICIENTE` | 409 | Saldo insuficiente para baixa | Revisar saldo e quantidade solicitada |
| `ERRO_BANCO` | 500 | Erro de persistência | Verificar banco e logs pelo `request_id` |

### Backend de faturamento

| Código | Status | Mensagem padrão | Ação operacional |
| --- | --- | --- | --- |
| `DADOS_INVALIDOS` | 400 | Dados inválidos para criar nota | Revisar estrutura do payload |
| `ITENS_INVALIDOS` | 400 | Nota sem item válido | Corrigir itens da nota |
| `NOTA_FECHADA` | 409 | Regra de status bloqueando ação | Aplicar ação compatível com o status |
| `NOTA_ABERTA` | 409 | Exclusão bloqueada para nota aberta | Cancelar/imprimir antes de excluir |
| `ESTOQUE_INDISPONIVEL` | 503 | Serviço de estoque indisponível | Reprocessar após restabelecer integração |
| `BAIXA_ESTOQUE_FALHOU` | 4xx/5xx | Falha ao baixar estoque | Consultar detalhes e corrigir causa raiz |

## Roteiro curto de troubleshooting

1. Capturar `request_id` retornado no erro.
2. Buscar o mesmo `request_id` nos logs do serviço.
3. Identificar código de erro e endpoint afetado.
4. Reproduzir cenário na rota correspondente.
5. Aplicar correção ou rollback conforme impacto.
