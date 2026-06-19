# FinanceOS — Pre-Launch Task List

> Gerado em 2026-06-17 a partir da análise completa do projeto.
> Base: TASKS.md (89/89 ✅), QA_REPORT.md (58/58 testes), STATUS_E_ROADMAP.md.

---

## Situação atual

O produto está **funcionalmente completo** (89/89 tasks, 58/58 testes QA). O que separa o código de um lançamento real é infraestrutura + polimento do diferencial.

---

## O que precisa ser feito antes de lançar

Organizado por prioridade, com foco no diferencial "finanças + investimentos num único lugar".

---

## 🔴 Bloqueadores — sem isso não tem lançamento

| # | Task | Detalhes |
|---|------|----------|
| B1 | **Deploy real (API + app)** | Hoje só funciona em Docker local. Escolher destino: Railway, Fly.io, VPS, etc. |
| B2 | **HTTPS / TLS** | `nginx.conf` não tem SSL. iOS bloqueia HTTP por padrão (App Transport Security). Pré-requisito do mobile. |
| B3 | **Base URL dinâmica no Flutter** | `api_client.dart` tem `localhost:8000` hardcoded. Sem isso o app mobile não conecta em produção. |
| B4 | **Rate limiting nos endpoints de auth** | `/auth/login`, `/register`, `/forgot-password` sem proteção contra brute force. |
| B5 | **Gestão de secrets** | `JWT_SECRET` e chaves de API em `.env` plaintext. Usar variáveis injetadas no CI/CD ou secrets manager. |
| B6 | **CORS de produção** | Hoje fixo em `localhost:3000`. Precisa do domínio real configurado. |
| B7 | **Backup do PostgreSQL** | Nenhuma estratégia definida. `pg_dump` agendado + restore testado. Sem isso dados de usuário real estão em risco. |

---

## 🟡 Essenciais para o diferencial "investimentos + finanças"

Essas são as features que fazem o usuário **perceber o valor** do produto.

| # | Task | Por que importa para o diferencial |
|---|------|------------------------------------|
| D1 | **Patrimônio líquido unificado no dashboard** | Conta bancária + carteira de investimentos + ativos customizados numa única métrica. É o número que o usuário quer ver todo dia. |
| D2 | **Gráfico de evolução patrimonial** | Histórico mensal de patrimônio total (não só saldo bancário). Diferencia de qualquer app de controle de gastos. |
| D3 | **Dividendos lançados automaticamente como receita** | Hoje a transação de dividendo existe no módulo de investimentos, mas não aparece no fluxo de caixa. Fechar esse loop. |
| D4 | **Capacidade de investimento mensal** | Widget no dashboard: "você tem R$ X sobrando este mês — isso representa Y% do seu salário". Liga gastos com investimentos diretamente. |
| D5 | **Relatório de IR de investimentos** | Para o mercado brasileiro isso é crítico. Day trade (tabela progressiva), swing trade (15%), FIIs (20%), dividendos isentos. Geração de DARF. |
| D6 | **Alerta de concentração de carteira** | Já existe a análise no backend (F11-05), mas precisa de notificação ativa quando um ativo passa de 30% do portfólio. |
| D7 | **Rentabilidade vs benchmark (CDI / IBOV / IPCA)** | Está parcialmente na tela de análise, mas o usuário precisa ver "minha carteira rendeu X% vs CDI de Y% no período" de forma clara. |
| D8 | **Metas vinculadas a portfólios** | Hoje metas e investimentos são silos separados. Permitir "essa meta será financiada com os rendimentos dessa carteira". |

---

## 🟠 Mobile — o canal de distribuição

| # | Task | Detalhes |
|---|------|----------|
| M1 | **Adicionar plataformas iOS/Android** | `flutter create --platforms=android,ios .` na pasta `apps/web`. O código já é 100% portável. |
| M2 | **Resolver base URL por plataforma** | Android emulador usa `10.0.2.2`, iOS usa `localhost`, produção usa o domínio real. |
| M3 | **Permissões nativas** | `AndroidManifest.xml` (INTERNET, file_picker), `Info.plist` (NSAppTransportSecurity em dev). |
| M4 | **Testar fluxos críticos em tela pequena** | Dashboard (fl_chart), formulário de transação, tabelas de investimentos. O `ResponsiveLayout` existe mas nunca rodou em device real. |
| M5 | **Identidade do app** | Bundle ID (`com.financeos.app`), ícone, splash screen, nome nas lojas. |
| M6 | **Assinatura e build de release** | Android: keystore + signingConfigs. iOS: Apple Developer ($99/ano) + provisioning profile. |
| M7 | **Push notifications** | Hoje são só in-app. FCM (Android) + APNs (iOS) ligados ao worker de notificações que já existe. Crítico para engajamento. |

---

## 🟢 Qualidade / Compliance

Importantes, mas não bloqueiam um soft launch.

| # | Task | Detalhes |
|---|------|----------|
| Q1 | **Exportação e exclusão de dados (LGPD)** | Obrigação legal no Brasil. Endpoint de export do usuário + exclusão completa de conta. |
| Q2 | **CI/CD (GitHub Actions)** | lint → `go test` → `flutter test` → build → deploy. Hoje nada disso é automatizado. |
| Q3 | **Paginação consistente em transações** | Com volume real de dados o endpoint degrada. Cursor-based pagination. |
| Q4 | **Monitoramento de erros (Sentry)** | Hoje o Zap só loga localmente. Sem visibilidade de erros em produção. |
| Q5 | **Retry de refresh token no Flutter** | Se o refresh falhar no meio de uma request, o usuário sofre logout fantasma sem motivo. |
| Q6 | **Billing real dos planos** | O middleware Free/Pro/Premium existe, mas sem Stripe/Pagar.me o produto não monetiza. |

---

## Ordem de execução sugerida

```
1. B1–B7  (infraestrutura)           ~1 semana
   └── pré-requisito: HTTPS antes de qualquer coisa

2. D1–D4  (diferencial visível)      ~1 semana
   └── isso é o que o usuário vai mostrar pra um amigo

3. M1–M6  (mobile)                   ~1–2 semanas
   └── depende do B2 (HTTPS) estar feito

4. D5     (IR de investimentos)      ~3–5 dias
   └── killer feature no Brasil — nenhum app de controle de gastos faz isso

5. Q1–Q6 + M7  (qualidade + push)    contínuo
```

---

## Por que o diferencial importa

O diferencial real do produto não está em controlar gastos — isso o Mobills e o GuiaBolso já fazem. Está em **conectar o fluxo de caixa diário com a construção de patrimônio**. As tasks D1–D5 são as que materializam isso na tela do usuário e justificam a escolha do FinanceOS sobre qualquer concorrente.
