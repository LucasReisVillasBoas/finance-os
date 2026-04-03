FinanceOS — Task Board
Legenda
🔲 Pendente | 🔄 Em andamento | ✅ Concluído | ❌ Bloqueado

FASE 1 — Fundação do Monorepo
✅ F1-01 - Criar estrutura de pastas do monorepo
financeos/
├── apps/
│   ├── api/           # Golang backend
│   └── web/           # Flutter web
├── packages/
│   ├── database/      # Migrations + seeds
│   └── shared/        # Tipos compartilhados
├── docker-compose.yml
├── .env.example
├── Makefile
└── CLAUDE.md
✅ F1-02 - Criar docker-compose.yml

PostgreSQL 16
Redis (cache + filas)
Adminer (UI de banco local)
Hot reload para API Go

✅ F1-03 - Criar Makefile com comandos principais

make dev (sobe tudo)
make migrate (roda migrations)
make seed (popula dados iniciais)
make test (roda todos os testes)
make build (build produção)

✅ F1-04 - Setup projeto Golang

go mod init
Estrutura Clean Architecture: cmd/, internal/domain/, internal/usecase/, internal/repository/, internal/handler/, internal/middleware/, pkg/
Instalar dependências: gin, pgx, redis, jwt, viper, zap, testify

✅ F1-05 - Setup Flutter web

flutter create com suporte web
Instalar pacotes: go_router, flutter_riverpod, dio, fl_chart, shared_preferences, flutter_secure_storage
Estrutura de pastas por feature: lib/features/, lib/core/, lib/shared/

✅ F1-06 - Criar sistema de migrations

golang-migrate
Migration 001: schema completo (usar schema.sql como base)
Migration 002: seed de categorias padrão
Script de rollback

✅ F1-07 - Configurar variáveis de ambiente

.env.example com todas as vars necessárias
Viper para carregar config em Go
Validação de vars obrigatórias no startup


FASE 2 — Autenticação
🔲 F2-01 - Endpoint POST /auth/register

Validar email único
Hash de senha com bcrypt
Criar usuário + portfolio padrão + dashboard padrão
Enviar email de verificação (mock local)
Retornar access_token + refresh_token

🔲 F2-02 - Endpoint POST /auth/login

Validar credenciais
Verificar se email foi verificado
Gerar JWT (access 15min + refresh 30 dias)
Salvar refresh_token no banco
Registrar last_login_at

🔲 F2-03 - Endpoint POST /auth/refresh

Validar refresh_token
Rotacionar token (revogar antigo, emitir novo)
Retornar novo par de tokens

🔲 F2-04 - Endpoint POST /auth/logout

Revogar refresh_token
Blacklist de JWT (Redis)

🔲 F2-05 - Endpoint POST /auth/forgot-password

Gerar token de reset
Enviar email (mock local)

🔲 F2-06 - Endpoint POST /auth/reset-password

Validar token
Atualizar senha
Revogar todos os refresh tokens do usuário

🔲 F2-07 - Middleware de autenticação JWT

Extrair e validar JWT em todas as rotas protegidas
Injetar user_id no contexto

🔲 F2-08 - Tela Flutter: Splash Screen

Logo + loading
Redireciona para login ou home conforme token

🔲 F2-09 - Tela Flutter: Login

Form email + senha
Link para cadastro e esqueci senha
Feedback de erro

🔲 F2-10 - Tela Flutter: Cadastro

Form nome, email, senha, confirmação
Validações client-side
Redirect para onboarding

🔲 F2-11 - Tela Flutter: Onboarding

3 passos: criar primeira conta bancária, definir salário, escolher categorias
Skip disponível


FASE 3 — Contas Bancárias
🔲 F3-01 - CRUD Accounts (API)

GET /accounts — listar contas do usuário
POST /accounts — criar conta
GET /accounts/:id — detalhe
PUT /accounts/:id — editar
DELETE /accounts/:id — desativar (soft delete)

🔲 F3-02 - GET /accounts/summary

Saldo total por conta
Saldo líquido total (excluindo cartão de crédito)
Patrimônio total

🔲 F3-03 - Tela Flutter: Lista de Contas

Cards por conta com saldo atual
Saldo total no topo
FAB para adicionar

🔲 F3-04 - Tela Flutter: Criar/Editar Conta

Tipo de conta (banco, cartão, carteira, etc.)
Nome, instituição, cor, ícone
Saldo inicial

🔲 F3-05 - Tela Flutter: Detalhe da Conta

Saldo atual + histórico de variação
Últimas transações da conta
Opção de editar/desativar


FASE 4 — Categorias
🔲 F4-01 - CRUD Categories (API)

GET /categories — listar (sistema + usuário)
POST /categories — criar customizada
PUT /categories/:id — editar (apenas as do usuário)
DELETE /categories/:id — desativar

🔲 F4-02 - Tela Flutter: Gerenciar Categorias

Lista separada por tipo (receita/despesa)
Indicador "padrão do sistema" vs "personalizada"
Criar nova com nome, ícone, cor


FASE 5 — Transações
🔲 F5-01 - CRUD Transactions (API)

GET /transactions — listar com filtros (data, categoria, conta, tipo)
POST /transactions — criar
GET /transactions/:id — detalhe
PUT /transactions/:id — editar
DELETE /transactions/:id — excluir

🔲 F5-02 - GET /transactions/summary

Total receitas / despesas / saldo no período
Agrupado por categoria
Comparativo com período anterior

🔲 F5-03 - Endpoint para transferência entre contas

POST /transactions/transfer
Cria 2 transações vinculadas (transfer_pair_id)
Debita conta origem, credita conta destino

🔲 F5-04 - Tela Flutter: Lista de Transações

Lista paginada com filtros
Agrupada por data
Indicador de tipo com cor

🔲 F5-05 - Tela Flutter: Criar Transação

Tipo (receita/despesa/transferência)
Valor com teclado numérico customizado
Data, categoria, conta, descrição, notas
Tags

🔲 F5-06 - Tela Flutter: Editar/Detalhe Transação

Todos os campos editáveis
Histórico de alterações

🔲 F5-07 - Tela Flutter: Filtros e Busca

Filtro por período, categoria, conta, tipo, tags
Busca por descrição


FASE 6 — Transações Recorrentes
🔲 F6-01 - CRUD Recurrences (API)

GET /recurrences
POST /recurrences
PUT /recurrences/:id
DELETE /recurrences/:id

🔲 F6-02 - Worker: processar recorrências diariamente

Verificar next_due_date
Se auto_launch, criar transação automaticamente
Notificar usuário se não for auto_launch

🔲 F6-03 - Tela Flutter: Gerenciar Recorrências

Lista com próxima data, valor, tipo
Toggle auto_launch


FASE 7 — Orçamento
🔲 F7-01 - CRUD Budgets (API)

GET /budgets — orçamentos do período atual
POST /budgets — criar
PUT /budgets/:id — editar
DELETE /budgets/:id — excluir

🔲 F7-02 - GET /budgets/progress

Para cada orçamento: previsto vs realizado vs percentual
Alertas (acima de threshold)

🔲 F7-03 - Tela Flutter: Orçamentos

Barra de progresso por categoria
Cores: verde (<70%), amarelo (70-90%), vermelho (>90%)
FAB para adicionar orçamento

🔲 F7-04 - Tela Flutter: Criar Orçamento

Selecionar categoria ou geral
Valor e período
Threshold de alerta


FASE 8 — Dashboard Principal
🔲 F8-01 - GET /dashboard/overview

Saldo total líquido
Receitas e despesas do mês
Top 5 categorias de gasto
Orçamentos em alerta
Patrimônio total (financeiro + investimentos + custom assets)

🔲 F8-02 - GET /dashboard/cashflow

Fluxo de caixa dos últimos 12 meses
Previsão próximos 3 meses (baseado em recorrências)

🔲 F8-03 - Tela Flutter: Home / Dashboard

Cards de saldo por conta
Gráfico de receitas vs despesas
Atalhos rápidos (+ transação, ver carteira)
Feed de últimas transações

🔲 F8-04 - Tela Flutter: Dashboard Customizável

Grid de widgets drag-and-drop
Adicionar/remover/reordenar widgets
Widgets: saldo, gráfico pizza categorias, fluxo de caixa, metas, carteira


FASE 9 — Investimentos: Core
🔲 F9-01 - CRUD Investment Portfolios (API)

GET /portfolios
POST /portfolios
PUT /portfolios/:id
DELETE /portfolios/:id

🔲 F9-02 - CRUD Holdings (API)

GET /portfolios/:id/holdings
POST /portfolios/:id/holdings
PUT /holdings/:id
DELETE /holdings/:id

🔲 F9-03 - CRUD Investment Transactions (API)

GET /holdings/:id/transactions
POST /holdings/:id/transactions — compra, venda, dividendo
DELETE /investment-transactions/:id

🔲 F9-04 - Lógica de cálculo de posição (Go)

Preço médio ponderado após compra/venda
Quantidade atualizada
P&L realizado ao vender
Total investido

🔲 F9-05 - CRUD Custom Assets (API)

GET /custom-assets
POST /custom-assets — imóvel, veículo, etc.
PUT /custom-assets/:id
DELETE /custom-assets/:id


FASE 10 — Investimentos: Preços e Cotações
🔲 F10-01 - Integração BRAPI (ações e FIIs BR)

GET /assets/search?q=PETR4 — busca ativo
Worker: atualizar preços B3 a cada 15min (horário de mercado)
Fallback se API indisponível

🔲 F10-02 - Integração Yahoo Finance (ações EUA, ETFs)

Busca e atualização de preços USD
Conversão BRL/USD via API de câmbio (AwesomeAPI)

🔲 F10-03 - Integração CoinGecko (cripto)

Busca por nome/ticker
Atualização de preços a cada 5min

🔲 F10-04 - Worker: calcular P&L de holdings

Para cada holding com asset vinculado
Recalcular current_value, unrealized_pnl, unrealized_pnl_pct
Rodar após cada atualização de preço

🔲 F10-05 - Cálculo de Renda Fixa (Go)

CDI diário via API Bacen
IPCA mensal via IBGE
Projeção de rendimento até vencimento
Cálculo de IR regressivo


FASE 11 — Investimentos: UI
🔲 F11-01 - Tela Flutter: Carteira

Valor total investido vs atual vs variação
Gráfico de alocação por classe (pizza)
Lista de holdings com P&L individual

🔲 F11-02 - Tela Flutter: Detalhe do Ativo

Gráfico de preço histórico (1d, 1m, 3m, 1a)
Posição atual, preço médio, quantidade
Histórico de operações

🔲 F11-03 - Tela Flutter: Adicionar Operação

Busca de ativo (ticker ou nome)
Tipo: compra, venda, dividendo
Quantidade, preço, taxas, data

🔲 F11-04 - Tela Flutter: Ativo Customizado

Form: nome, classe, valor atual, valor de compra, renda mensal
Imóveis: endereço, aluguel mensal

🔲 F11-05 - Tela Flutter: Análise de Carteira

Diversificação por setor, classe, moeda
Concentração (alertar >30% em ativo único)
Comparativo com benchmark (IBOV, CDI)


FASE 12 — Metas
🔲 F12-01 - CRUD Goals (API)

GET /goals
POST /goals
PUT /goals/:id
DELETE /goals/:id
POST /goals/:id/contribute — adicionar aporte

🔲 F12-02 - GET /goals/projections

Estimativa de quando cada meta será atingida
Com e sem aportes mensais

🔲 F12-03 - Tela Flutter: Metas

Cards com barra de progresso
Estimativa de conclusão
Botão de aporte


FASE 13 — Importação de Dados
🔲 F13-01 - Importação OFX (API)

POST /imports/ofx — upload de arquivo
Parser de OFX em Go
Detecção de duplicatas via import_id
Pré-categorização automática por histórico

🔲 F13-02 - Importação CSV (API)

POST /imports/csv
Mapeamento de colunas configurável
Preview antes de confirmar

🔲 F13-03 - Tela Flutter: Importar Extrato

Upload de arquivo OFX ou CSV
Mapeamento de campos
Preview com opção de editar antes de importar
Relatório de resultado (importados, duplicatas, erros)


FASE 14 — WhatsApp Bot
🔲 F14-01 - Setup Evolution API (Docker)

Adicionar ao docker-compose.yml
Webhook configurado para API Go
QR Code de conexão

🔲 F14-02 - Endpoint POST /webhooks/whatsapp

Receber mensagens da Evolution API
Identificar usuário pelo número
Roteamento para state machine

🔲 F14-03 - State machine do bot (Go)

Estado idle: aguardar comando
Processar "gastei X reais em Y" → criar transação
Processar "quanto gastei esse mês?" → responder
Confirmar antes de salvar
Timeout de sessão (15min)

🔲 F14-04 - Integração Claude API no bot

Interpretar mensagem em linguagem natural
Extrair: valor, categoria sugerida, descrição, data
Confidence score — se baixo, pedir confirmação

🔲 F14-05 - Comandos do bot

"resumo" → saldo e gastos do mês
"gastei X" → lançar despesa
"recebi X" → lançar receita
"quanto gastei com alimentação?" → consulta por categoria
"carteira" → resumo de investimentos

🔲 F14-06 - Tela Flutter: Configurar WhatsApp Bot

Vincular número
Ver histórico de mensagens processadas
Ativar/desativar


FASE 15 — IA Features
🔲 F15-01 - Categorização automática (API)

Worker: categorizar transações sem categoria via Claude API
Usar histórico do usuário como contexto
Salvar ai_categorized=true + ai_confidence
Só ativar para plano Pro+

🔲 F15-02 - Previsão de gastos (API)

GET /ai/spending-forecast
Analisar padrão dos últimos 6 meses
Projetar próximos 3 meses por categoria
Cache de 24h no Redis

🔲 F15-03 - Alertas inteligentes (Worker)

"Você gasta X% mais em dezembro"
"Categoria Y aumentou Z% esse mês"
"Você vai estourar o orçamento de X"
Gerar via Claude API, salvar em ai_suggestions

🔲 F15-04 - Análise de carteira IA

GET /ai/portfolio-analysis
Identificar concentração excessiva
Sugerir rebalanceamento
Comparar perfil de risco

🔲 F15-05 - Assistente conversacional (Flutter)

Chat widget no app
Contexto: saldo, gastos do mês, carteira
Responder perguntas financeiras personalizadas


FASE 16 — Notificações
🔲 F16-01 - Sistema de notificações (API)

GET /notifications — listar
PUT /notifications/:id/read — marcar como lida
DELETE /notifications — limpar todas

🔲 F16-02 - Worker: disparar notificações

Checar orçamentos acima do threshold
Metas próximas de vencer
Recorrências vencendo amanhã
Resumo semanal (domingo)
Relatório mensal (dia 1)

🔲 F16-03 - Email: templates

Resumo semanal
Alerta de orçamento
Relatório mensal
Usar templates HTML simples

🔲 F16-04 - Tela Flutter: Central de Notificações

Lista de notificações com ícone por tipo
Marcar como lida / limpar
Badge no ícone do app

🔲 F16-05 - Tela Flutter: Preferências de Notificação

Toggles por tipo e canal
Configurar horário de resumo


FASE 17 — Multi-usuário / Família
🔲 F17-01 - CRUD Family Groups (API)

POST /family — criar grupo
GET /family — meu grupo
POST /family/invite — gerar convite
POST /family/join — entrar com código
DELETE /family/members/:id — remover membro

🔲 F17-02 - Visão consolidada família (API)

GET /family/dashboard — soma de todos os membros
Permissões por membro (view_transactions, view_investments)

🔲 F17-03 - Tela Flutter: Família

Criar/entrar em grupo
Lista de membros com permissões
Toggle de contas compartilhadas


FASE 18 — Planos e Assinatura
🔲 F18-01 - Middleware de plano (Go)

Verificar plano do usuário em rotas premium
Retornar 402 Payment Required com mensagem clara

🔲 F18-02 - Feature flags por plano
Free:    contas (até 2), transações (ilimitado), categorias padrão, dashboard básico
Pro:     contas ilimitadas, investimentos, IA básica, WhatsApp bot, importação
Premium: família, IA completa, dashboards customizados, alertas avançados
🔲 F18-03 - Tela Flutter: Planos e Upgrade

Comparativo de planos em cards
CTA para upgrade (link externo por enquanto)
Mostrar o que está bloqueado com cadeado


FASE 19 — Polimento e Qualidade
🔲 F19-01 - Tela Flutter: Perfil do Usuário

Foto, nome, email
Trocar senha
Fuso horário e moeda
Excluir conta

🔲 F19-02 - Tela Flutter: Configurações

Preferências de notificação
Categorias customizadas
Campos customizados
Sobre o app / versão

🔲 F19-03 - Tratamento global de erros (Flutter)

Interceptor Dio para erros de API
Snackbars padronizados
Tela de erro genérica com retry

🔲 F19-04 - Loading states e skeleton screens

Skeleton para todas as listas
Shimmer effect enquanto carrega

🔲 F19-05 - Testes de integração (API)

Testes end-to-end dos principais fluxos
Auth, transações, investimentos, orçamento

🔲 F19-06 - Testes de widget (Flutter)

Telas principais: home, transações, carteira

🔲 F19-07 - Responsividade web (Flutter)

Breakpoints: mobile (<768px), tablet (768-1200px), desktop (>1200px)
Layout adaptado para cada tamanho


FASE 20 — Deploy Local
🔲 F20-01 - Build Flutter web para produção

flutter build web
Servir com nginx no Docker

🔲 F20-02 - Build Go para produção

Compilar binário otimizado
Dockerfile multi-stage

🔲 F20-03 - docker-compose.prod.yml

Sem volumes de desenvolvimento
Variáveis de ambiente via .env
Restart policies

🔲 F20-04 - Script de setup inicial

Rodar migrations
Seed de categorias padrão
Criar usuário admin de teste


Contador de Tasks
Total: 89 tasks | ✅ 7 | 🔄 0 | 🔲 82 | ❌ 0