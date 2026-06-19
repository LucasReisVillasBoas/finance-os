# FinanceOS — CLAUDE.md Master

## Missão
Você é o orquestrador do projeto FinanceOS. Seu objetivo é construir
o aplicativo completo de forma totalmente autônoma, task por task,
sem intervenção humana, até todas as tasks em TASKS.md estarem com status ✅.

> **Estado atual:** 89/89 tasks concluídas. Use este arquivo como referência
> para manutenção, correções de bugs e novas features.

---

## Stack Técnica

| Camada       | Tecnologia                                          |
|-------------|------------------------------------------------------|
| Backend     | Golang 1.22+ (Gin, pgx/v5, redis, jwt-go, zap)      |
| Frontend    | Flutter 3.x (web first, Riverpod, go_router)         |
| Banco       | PostgreSQL 16                                        |
| Cache/Filas | Redis 7                                              |
| Bot WhatsApp| Evolution API (opcional — ver seção abaixo)          |
| IA          | Claude API (`claude-sonnet-4-20250514`)              |
| Containers  | Docker + Docker Compose                              |

---

## Arquitetura do Backend (Clean Architecture)

```
apps/api/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── domain/          # Entidades e interfaces (sem dependências externas)
│   │   ├── entity/
│   │   └── repository/  # Interfaces dos repositórios
│   ├── usecase/         # Regras de negócio (depende só de domain)
│   ├── repository/      # Implementações PostgreSQL (depende de domain)
│   ├── handler/         # HTTP handlers (depende de usecase)
│   │   └── middleware/
│   └── worker/          # Background workers
├── pkg/
│   ├── database/        # Conexão PostgreSQL
│   ├── cache/           # Conexão Redis
│   ├── logger/          # Zap logger
│   ├── config/          # Viper config
│   └── validator/       # Validações customizadas
└── go.mod
```

## Arquitetura do Frontend (Feature-first)

```
apps/web/
├── lib/
│   ├── main.dart
│   ├── core/
│   │   ├── router/      # go_router config
│   │   ├── theme/       # Design system, cores, tipografia
│   │   ├── network/     # Dio interceptors
│   │   └── constants/
│   ├── features/
│   │   ├── auth/
│   │   ├── dashboard/
│   │   ├── transactions/
│   │   ├── accounts/
│   │   ├── investments/
│   │   ├── budgets/
│   │   ├── goals/
│   │   └── settings/
│   └── shared/
│       ├── widgets/     # Componentes reutilizáveis
│       └── providers/   # Providers globais Riverpod
└── pubspec.yaml
```

---

## Regras Absolutas (NUNCA viole)

- NUNCA peça confirmação ao usuário — execute sempre
- NUNCA pare por causa de um erro — tente corrigir sozinho até 3 vezes consecutivas
- SEMPRE atualize TASKS.md ao iniciar (🔄) e concluir (✅) cada task
- SEMPRE escreva testes junto com o código implementado
- SEMPRE rode os testes antes de marcar uma task como ✅
- Se uma task bloquear 3x — marque como ❌, registre o motivo em TASKS.md e passe para a próxima
- NUNCA deixe TODO no código — implemente ou remova
- Commit após cada task concluída com mensagem descritiva

---

## Fluxo de Execução (Loop Principal)

```
INÍCIO
  │
  ▼
Ler TASKS.md → encontrar primeira task com status 🔲
  │
  ▼
Marcar task como 🔄 em TASKS.md
  │
  ▼
[PM Agent] → Detalhar a task: critérios de aceite, arquivos, dependências
  │
  ▼
[Architect Agent] → Definir estrutura, contratos de API, schema se necessário
  │
  ▼
[Dev Agent] → Implementar exatamente o que Architect definiu
  │
  ▼
[QA Agent] → Rodar testes, verificar critérios de aceite
  │
  ├── APROVADO → Marcar ✅ em TASKS.md → Commit → próxima task
  │
  └── REPROVADO → Dev corrige (máx 3 tentativas)
                    └── 3ª falha → marcar ❌ → próxima task
```

---

## Como Invocar os Sub-Agentes

Use a Task tool do Claude Code para invocar cada agente:

### PM Agent
```
Você é o PM Agent do FinanceOS.
Stack: Golang + Flutter + PostgreSQL.
Task atual: [NOME DA TASK]
Descrição breve: [DESCRIÇÃO]

Retorne SOMENTE um JSON com:
{
  "task_id": "...",
  "detailed_description": "...",
  "acceptance_criteria": ["...", "..."],
  "files_to_create": ["..."],
  "files_to_modify": ["..."],
  "api_contracts": [{"method": "POST", "path": "/...", "request": {}, "response": {}}],
  "dependencies": ["task_ids que devem estar prontas antes"]
}
```

### Architect Agent
```
Você é o Architect Agent do FinanceOS.
Stack: Golang Clean Architecture + Flutter Riverpod + PostgreSQL.

Receba a definição do PM e retorne:
1. Estrutura de arquivos a criar (paths completos)
2. Interfaces e types necessários (Go structs / Dart classes)
3. Contrato de API detalhado (se aplicável)
4. Schema de banco (se aplicável)
5. Fluxo de dados entre camadas

Seja específico o suficiente para o Dev Agent implementar sem perguntas.
```

### Dev Agent
```
Você é o Dev Agent do FinanceOS.
Implemente exatamente o que o Architect definiu.

Padrões obrigatórios:
- Go: Clean Architecture, error wrapping com fmt.Errorf, context propagation
- Flutter: Riverpod para estado, Repository pattern para API calls
- Testes: table-driven tests em Go, widget tests em Flutter
- Nunca hardcode credenciais
- Sempre validate inputs no handler (Go) e no form (Flutter)
```

### QA Agent
```
Você é o QA Agent do FinanceOS.

Para cada task implementada:
1. Execute: go test ./... (no diretório apps/api)
2. Execute: flutter test (no diretório apps/web)
3. Verifique manualmente os critérios de aceite do PM
4. Tente casos de borda: valores nulos, strings vazias, IDs inválidos

Retorne: APROVADO ou REPROVADO
Se REPROVADO, liste bugs com: arquivo, linha, descrição, reprodução
```

---

## Padrões de Código

### Go — Estrutura de um Handler
```go
// handler/account_handler.go
type AccountHandler struct {
    usecase usecase.AccountUseCase
    logger  *zap.Logger
}

func (h *AccountHandler) Create(c *gin.Context) {
    var req dto.CreateAccountRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    userID := c.GetString("user_id")
    account, err := h.usecase.Create(c.Request.Context(), userID, req)
    if err != nil {
        h.logger.Error("create account", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    c.JSON(http.StatusCreated, account)
}
```

### Flutter — Estrutura de um Provider
```dart
// features/accounts/providers/accounts_provider.dart
@riverpod
class AccountsNotifier extends _$AccountsNotifier {
  @override
  Future<List<Account>> build() => ref.read(accountRepositoryProvider).getAll();

  Future<void> create(CreateAccountDto dto) async {
    await ref.read(accountRepositoryProvider).create(dto);
    ref.invalidateSelf();
  }
}
```

### Flutter — Estrutura de uma Tela
```dart
// features/accounts/screens/accounts_screen.dart
class AccountsScreen extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final accounts = ref.watch(accountsNotifierProvider);
    return accounts.when(
      data: (data) => AccountsList(accounts: data),
      loading: () => const AccountsSkeleton(),
      error: (e, _) => ErrorView(onRetry: () => ref.invalidate(accountsNotifierProvider)),
    );
  }
}
```

---

## API Response Pattern

Todas as respostas da API devem seguir:

```json
// Sucesso
{"data": {...}, "meta": {"page": 1, "total": 100}}

// Erro
{"error": {"code": "INVALID_INPUT", "message": "...", "details": {}}}
```

---

## Variáveis de Ambiente Necessárias

```env
# Database
# ATENÇÃO: porta 5434 externamente (mapeada de 5432 no container)
DATABASE_URL=postgresql://financeos:financeos@localhost:5434/financeos?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=your-secret-here
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=720h

# Claude API
ANTHROPIC_API_KEY=your-key-here

# Evolution API (WhatsApp) — OPCIONAL, projeto funciona sem isso
EVOLUTION_API_URL=http://localhost:8081
EVOLUTION_API_KEY=your-key-here

# App
APP_ENV=development
APP_PORT=8000
LOG_LEVEL=debug
```

---

## Workflow de Desenvolvimento

### Modo simples (todos os serviços no Docker, sem hot reload)
```bash
make dev        # sobe tudo
make dev-down   # derruba tudo
make dev-logs   # ver logs
```

### Modo hot reload (recomendado para desenvolvimento Flutter)
```bash
# Terminal 1 — backend (postgres, redis, api, adminer)
make dev-backend

# Terminal 2 — Flutter no host com hot reload
make web
# Acesse: http://localhost:3000
# 'r' → hot reload  |  'R' → hot restart  |  'q' → encerra
```

### Outros comandos úteis
```bash
make migrate      # rodar migrations
make migrate-down # rollback última migration
make seed         # popular dados iniciais
make test         # rodar todos os testes
make build        # build produção
make lint         # go vet + dart analyze
make fmt          # formatar código
make deps         # instalar dependências
```

---

## Serviços e Portas

| Serviço   | URL / Porta                  | Notas                          |
|-----------|------------------------------|--------------------------------|
| API Go    | http://localhost:8000        | Container + host               |
| Flutter   | http://localhost:3000        | Container (sem HR) ou host (HR)|
| Adminer   | http://localhost:8080        | UI para PostgreSQL             |
| PostgreSQL| localhost:**5434**           | Mapeado de 5432 no container   |
| Redis     | localhost:6379               |                                |
| Evolution | http://localhost:8081        | Opcional — WhatsApp Bot        |

---

## Evolution API (WhatsApp Bot) — Opcional

O serviço `financeos_evolution` no docker-compose é **opcional**. O projeto
funciona completamente sem ele. Para subir sem o bot:

```bash
docker-compose up -d postgres redis api adminer
```

Para subir com o bot:
```bash
docker-compose up -d
```

---

## Gotchas e Lições Aprendidas

### Go
- **Nil slices serializam como `null` em JSON** — sempre inicialize com slice vazio:
  ```go
  if results == nil {
      results = []*entity.Something{}
  }
  ```
- **`golang-migrate` instalado em `~/go/bin`** que pode não estar no PATH.
  O Makefile usa `$(HOME)/go/bin/migrate` como fallback automaticamente.
- **PostgreSQL local precisa de `sslmode=disable`** — inclua sempre na DATABASE_URL.

### Flutter / Dart
- **`toIso8601String()` não inclui timezone** — use sempre `.toUtc().toIso8601String()`
  para evitar erros de parsing no Go (`cannot parse "" as "Z07:00"`).
- **Casts inseguros de List** — nunca use `data['data'] as List<dynamic>`.
  Use `(data['data'] as List<dynamic>?) ?? []` para evitar TypeError em listas nulas.
- **Interceptor Dio assíncrono sem try/catch** causa loading infinito — sempre
  envolva `storage.read()` em try/catch com timeout.
- **Hierarquia de investimentos**: Portfolio → Holding → InvestmentTransaction.
  Para criar uma operação, é necessário criar um Holding primeiro via
  `POST /portfolios/:id/holdings`, depois usar o holding ID em
  `POST /holdings/:id/transactions`.

### Cotações (APIs externas, gratuitas)
- **Busca de ações (B3)**: usa a [BRAPI](https://brapi.dev) (`pkg/brapi`).
  Hoje o brapi.dev exige um token gratuito para os endpoints de cotação —
  configure `BRAPI_TOKEN` no `.env`. Sem token a busca ainda funciona, mas
  os preços podem vir zerados. O token é anexado como `&token=` nas chamadas.
- **Cotações de moedas (USD/EUR)**: usa a
  [AwesomeAPI](https://docs.awesomeapi.com.br) (`pkg/currency`), gratuita e
  sem chave. Endpoint: `GET /api/v1/quotes/currencies` (cache Redis de 10min,
  chave `quotes:currencies`). Pares padrão: `USD-BRL`, `EUR-BRL`.

---

## Ordem de Execução das Fases

Execute SEMPRE nesta ordem — cada fase depende da anterior:

| Fase | Nome                    | Status |
|------|-------------------------|--------|
| 1    | Fundação                | ✅     |
| 2    | Auth                    | ✅     |
| 3    | Contas                  | ✅     |
| 4    | Categorias              | ✅     |
| 5    | Transações              | ✅     |
| 6    | Recorrências            | ✅     |
| 7    | Orçamento               | ✅     |
| 8    | Dashboard               | ✅     |
| 9    | Investimentos Core      | ✅     |
| 10   | Cotações e Preços       | ✅     |
| 11   | UI Investimentos        | ✅     |
| 12   | Metas                   | ✅     |
| 13   | Importação              | ✅     |
| 14   | WhatsApp Bot            | ✅     |
| 15   | IA Features             | ✅     |
| 16   | Notificações            | ✅     |
| 17   | Multi-usuário / Família | ✅     |
| 18   | Planos                  | ✅     |
| 19   | Polimento               | ✅     |
| 20   | Deploy Local            | ✅     |

---

## Comportamento em Caso de Erro

| Tipo de erro                  | Ação                                                    |
|-------------------------------|---------------------------------------------------------|
| Erro de compilação Go         | Leia o erro, corrija, recompile                         |
| Erro de teste                 | Leia o stack trace, corrija o código ou o teste         |
| Dependência não encontrada    | `go get` ou `flutter pub add`                           |
| Porta em uso                  | Use a próxima disponível                                |
| Migração falhou               | `migrate down`, corrija o SQL, rode novamente           |
| API externa indisponível      | Implemente mock/stub e continue                         |

Se após 3 tentativas o erro persiste:
1. Documente o erro no TASKS.md junto à task com ❌
2. Continue para a próxima task
3. NÃO pare a execução

---

## Comando para Iniciar

```bash
claude --dangerously-skip-permissions \
  "Leia o CLAUDE.md completamente. Em seguida leia o TASKS.md. \
   Execute o loop de orquestração: para cada task 🔲 em ordem, \
   invoque os agentes PM → Architect → Dev → QA, atualize o status \
   no TASKS.md, faça commit, e passe para a próxima. \
   Continue sem parar até todas as tasks estarem ✅ ou ❌. \
   Nunca peça confirmação. Nunca pare por erros recuperáveis."
```
