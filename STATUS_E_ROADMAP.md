# FinanceOS — Status do Projeto e Roadmap (v0 + Mobile)

> Gerado em 2026-06-13 a partir de exploração completa do código.
> Referências: `TASKS.md` (89/89 ✅), `QA_REPORT.md` (58/58 testes passando), `CLAUDE.md`.

---

## 1. Visão Geral

| Camada | Tecnologia | Estado |
|--------|-----------|--------|
| Backend | Go 1.22+ (Gin, pgx/v5, Redis, JWT, Zap) — Clean Architecture | ✅ Completo |
| Frontend | Flutter 3.x (Riverpod, go_router, Dio) — web-first | ✅ Completo (só web) |
| Banco | PostgreSQL 16 (2 migrations: schema completo + seed de categorias) | ✅ |
| Cache/Filas | Redis 7 | ✅ |
| IA | Claude API (chat, forecast, categorização automática) | ✅ (requer `ANTHROPIC_API_KEY`) |
| WhatsApp | Evolution API (opcional) | ✅ |
| Infra | Docker Compose (dev + prod), Nginx, Makefile | ✅ Deploy local |
| Mobile (Android/iOS) | Flutter (mesmo projeto, pastas `android/` e `ios/` geradas) | 🔄 Iniciado — falta validar em emulador e assinar |
| CI/CD | — | ❌ Não existe |

---

## 2. Funcionalidades Implementadas

### Backend (`apps/api`) — 12 domínios, ~50 endpoints em `/api/v1`

- **Auth** — register, login, refresh com rotation, logout, forgot/reset password
- **Contas** — CRUD + summary de saldos
- **Categorias** — 15 de sistema + customizadas por usuário
- **Transações** — CRUD, transferências entre contas (com `transfer_pair_id`), summary por período/categoria, filtros, tags
- **Recorrências** — CRUD + worker diário de auto-lançamento
- **Orçamentos** — CRUD, progresso, alertas por threshold
- **Dashboard** — overview (saldo, receitas/despesas, top categorias, patrimônio) e cashflow 12 meses com previsão
- **Investimentos** — Portfolios → Holdings → Transações, custom assets (imóveis etc.), busca de ativos, P&L, worker de preços (BRAPI/Yahoo/CoinGecko)
- **Metas** — CRUD, aportes, projeções (meses até a meta, data estimada)
- **Importação** — OFX e CSV com preview e detecção de duplicatas (Pro+)
- **IA** — chat, spending forecast, portfolio analysis (Pro+), categorização automática via worker
- **Notificações** — in-app, worker de alertas (orçamento, metas, recorrências, resumo semanal)
- **Família/Multi-usuário** — grupos, convites, permissões, dashboard consolidado
- **Planos** — middleware Free/Pro/Premium com feature gating

Pontos de entrada: `apps/api/internal/handler/router.go` (rotas) e `apps/api/cmd/server/main.go` (4 workers: recorrência, preços, IA, notificações).

**Testes backend:** 14 arquivos `*_test.go` (~5.000 linhas entre usecases e handlers) + 1 integration test pequeno + suíte QA externa (`qa/qa_suite.js`, 58/58).

### Frontend (`apps/web`) — 32 telas em 8 features

- **auth** (splash, login, register, onboarding) · **dashboard** (home, config)
- **accounts** (lista, form, detalhe) · **transactions** (lista, form, detalhe, filtros, recorrências)
- **budgets** (lista, form) · **goals** (lista, form, aporte)
- **investments** (portfolio, holding, forms, análise) · **settings** (perfil, categorias, notificações, preferências, planos, família, importação, IA, WhatsApp)

Infra do app: tokens em `flutter_secure_storage` com refresh automático (`lib/shared/providers/auth_provider.dart`), interceptors Dio (`lib/core/network/api_client.dart`), design responsivo com breakpoints mobile/tablet/desktop (`lib/core/constants/breakpoints.dart`, `lib/shared/widgets/responsive_layout.dart`).

**Testes frontend:** apenas 1 arquivo de widget tests (3 telas) — ponto fraco.

---

## 3. O Que Falta — Separado por Prioridade para uma v0 Completa

### 🔴 Bloqueadores de v0 (sem isso não dá para lançar)

| # | Item | Detalhe |
|---|------|---------|
| 1 | **CI/CD** | Não existe `.github/workflows/`. Criar pipeline: lint → `go test` + `flutter test` → build → deploy. |
| 2 | **TLS/HTTPS** | `nginx.conf` não tem SSL. Configurar certificado (Let's Encrypt) ou colocar atrás de um proxy gerenciado. |
| 3 | **Gestão de secrets** | `JWT_SECRET` e chaves vivem em `.env` plaintext. Usar secrets manager ou variáveis injetadas no deploy; rotacionar o secret atual. |
| 4 | **Rate limiting / proteção de auth** | Sem rate limit nos endpoints de login/register/forgot-password — vulnerável a brute force. Middleware com Redis resolve. |
| 5 | **Deploy real** | Hoje só existe "deploy local" (Fase 20). Escolher destino (VPS, Fly.io, Railway, Cloud Run) e publicar API + web. |
| 6 | **CORS de produção** | `CORS_ORIGINS` configurado só para `localhost:3000`; parametrizar para o domínio real. |
| 7 | **Backups do PostgreSQL** | Nenhuma estratégia. Script de `pg_dump` agendado + procedimento de restore testado. |

### 🟡 Importantes para uma v0 sólida (lançável sem, mas arriscado)

| # | Item | Detalhe |
|---|------|---------|
| 8 | **Testes E2E de fluxo completo** | Existe smoke test Playwright na suíte QA, mas nada automatizado em CI cobrindo register → transação → dashboard. |
| 9 | **Paginação em transações** | Sem cursor/limit consistente — degrada com milhares de registros. |
| 10 | **Monitoramento de erros** | Zap loga local apenas. Integrar Sentry (Go + Flutter) e health checks com alerta. |
| 11 | **Retry no refresh de token (Flutter)** | Se o refresh falhar no meio de uma request não há retry automático — risco de logout fantasma. |
| 12 | **Export/exclusão de dados (LGPD)** | Sem endpoint de exportação de dados do usuário nem exclusão de conta completa. |
| 13 | **Mais testes de widget/golden no Flutter** | 1 arquivo cobrindo 3 telas de 32. |

### 🟢 Pós-v0 (evolução)

- **Push notifications** (FCM/APNs) — hoje só in-app; essencial quando o mobile existir
- **Multi-moeda** — hardcoded BRL
- **Observabilidade** — Prometheus/OpenTelemetry
- **Billing real dos planos** — middleware existe, mas sem gateway de pagamento (Stripe/Pagar.me)
- **Feature flags** para rollout gradual

---

## 4. Mobile (iOS + Android) a Partir do Mesmo Projeto

### Boa notícia: o código já é ~100% portável

A auditoria de dependências não encontrou nenhum bloqueador:

- ✅ Nenhum uso de `dart:html` ou APIs web-only no código
- ✅ `flutter_riverpod`, `go_router`, `dio`, `fl_chart`, `intl`, `shared_preferences`, `flutter_secure_storage` — todos suportam Android/iOS/Web
- ⚠️ `file_picker` (usado na importação OFX/CSV) funciona em mobile, mas exige permissões por plataforma
- ⚠️ `apps/web/lib/core/network/api_client.dart:5` — base URL via `--dart-define` com default `http://localhost:8000`, que **não funciona** em device/emulador

O app é um único projeto Flutter: **não crie outro app** — apenas adicione as plataformas à pasta `apps/web` (vale considerar renomeá-la para `apps/app` no futuro, já que deixará de ser só web).

### Passo a passo (ordem recomendada)

> **Atualização 2026-06-13:** os passos 1–3 já foram executados neste repositório:
> pastas `android/` e `ios/` criadas, base URL por plataforma em `api_client.dart`
> (web/iOS → `localhost`, emulador Android → `10.0.2.2`, release → `--dart-define`),
> `applicationId`/bundle ID definidos como `com.financeos.app`, permissão INTERNET,
> cleartext HTTP somente em debug (Android) e exceção ATS de rede local (iOS).
> Continue do **Passo 4** (validar em emulador).

**Passo 1 — Adicionar as plataformas (5 min):**
```bash
cd apps/web
flutter create --platforms=android,ios --org com.financeos --project-name financeos_web .
```
Isso gera as pastas `android/` e `ios/` sem tocar no `lib/`.

**Passo 2 — Resolver a base URL por plataforma (crítico):**
```dart
// lib/core/network/api_client.dart
import 'dart:io' show Platform;
import 'package:flutter/foundation.dart' show kIsWeb;

const _envUrl = String.fromEnvironment('API_BASE_URL');

String get baseUrl {
  if (_envUrl.isNotEmpty) return _envUrl;
  if (kIsWeb) return 'http://localhost:8000';
  if (Platform.isAndroid) return 'http://10.0.2.2:8000'; // emulador Android → host
  return 'http://localhost:8000'; // simulador iOS enxerga o host direto
}
```
Em produção, sempre buildar com `--dart-define=API_BASE_URL=https://api.seudominio.com`. **Dependência:** o item 2 da v0 (HTTPS) é pré-requisito do mobile — iOS bloqueia HTTP por padrão (App Transport Security) e Android 9+ bloqueia cleartext. Para dev, libere exceções apenas em debug.

**Passo 3 — Permissões e configs nativas:**
- `android/app/src/main/AndroidManifest.xml`: permissão `INTERNET` (o `flutter create` já inclui em debug; confirmar no release) e, para o `file_picker`, nada extra em Android 13+ (usa o photo picker/SAF)
- `ios/Runner/Info.plist`: nada obrigatório para `file_picker` de documentos; revisar `NSAppTransportSecurity` apenas se precisar de HTTP em dev
- `flutter_secure_storage` no Android: ativar `EncryptedSharedPreferences`; no iOS funciona via Keychain sem config

**Passo 4 — Rodar e validar em emulador:**
```bash
flutter run -d <android-emulator|iphone-simulator> --dart-define=API_BASE_URL=http://10.0.2.2:8000
```
Validar os fluxos com maior risco em telas pequenas: dashboard (gráficos `fl_chart`), formulário de transação, importação (file_picker) e tabelas de investimentos. O `ResponsiveLayout` já existe, mas nunca foi testado fora do navegador.

**Passo 5 — Identidade e assinatura:**
- Android: definir `applicationId` (`com.financeos.app`), gerar keystore, configurar `signingConfigs` no `build.gradle`
- iOS: bundle ID no Xcode, conta Apple Developer (US$ 99/ano), certificado e provisioning profile — **iOS exige macOS com Xcode para buildar**; sem Mac, usar Codemagic/GitHub Actions com runner macOS

**Passo 6 — Build de release:**
```bash
flutter build appbundle --dart-define=API_BASE_URL=https://api.seudominio.com   # Android (Play Store)
flutter build ipa --dart-define=API_BASE_URL=https://api.seudominio.com        # iOS (App Store, requer macOS)
```

**Passo 7 — Pós-port (backlog mobile):**
- Push notifications (FCM + APNs) ligadas ao worker de notificações existente
- Biometria para login (local_auth) reaproveitando o refresh token guardado
- Deep links (go_router já suporta) para convites de família
- Adicionar os builds Android/iOS ao pipeline de CI

### Estimativa

| Fase | Esforço |
|------|---------|
| Passos 1–4 (app rodando em emulador iOS/Android) | 1–3 dias |
| Passos 5–6 (builds assinados nas lojas, contas criadas) | 3–5 dias + revisão das lojas |
| Ajustes de UX mobile descobertos no passo 4 | 2–5 dias (depende do que aparecer) |

---

## 5. Resumo Executivo

- **O produto está funcionalmente completo** para uso em dev: 89/89 tasks, 58/58 testes QA, todos os fluxos principais operando (auth, contas, transações, orçamentos, investimentos, metas, IA, família, planos).
- **O que separa o projeto de uma v0 publicada não é feature, é operação**: CI/CD, HTTPS, secrets, rate limiting, deploy real, backups (itens 1–7 da seção 3). Estimativa: ~2–3 semanas.
- **O mobile é o passo mais barato do roadmap**: o código Flutter já é portável, sem dependências web-only. Um `flutter create --platforms=android,ios` + ajuste da base URL coloca o app rodando em emulador em poucos dias. O único pré-requisito real é a API publicada em HTTPS.
- **Ordem sugerida**: fechar os bloqueadores da v0 (especialmente HTTPS/deploy) → portar para Android/iOS → push notifications → billing dos planos.
