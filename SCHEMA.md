-- ============================================================
-- FINANCEOS — Schema PostgreSQL Completo
-- ============================================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- ENUMS
-- ============================================================

CREATE TYPE plan_type AS ENUM ('free', 'pro', 'premium');
CREATE TYPE transaction_type AS ENUM ('income', 'expense', 'transfer');
CREATE TYPE account_type AS ENUM ('checking', 'savings', 'credit_card', 'investment', 'cash', 'custom');
CREATE TYPE asset_class AS ENUM ('stock_br', 'stock_us', 'fii', 'etf_br', 'etf_us', 'crypto', 'fixed_income', 'real_estate', 'vehicle', 'private_pension', 'custom');
CREATE TYPE fixed_income_type AS ENUM ('cdb', 'lci', 'lca', 'lc', 'cri', 'cra', 'debenture', 'tesouro_selic', 'tesouro_ipca', 'tesouro_prefixado', 'other');
CREATE TYPE indexer_type AS ENUM ('cdi', 'ipca', 'igpm', 'selic', 'prefixado', 'other');
CREATE TYPE notification_channel AS ENUM ('push', 'email', 'whatsapp');
CREATE TYPE notification_type AS ENUM ('budget_alert', 'goal_reached', 'investment_update', 'bill_due', 'weekly_summary', 'monthly_report', 'custom');
CREATE TYPE family_role AS ENUM ('owner', 'admin', 'member', 'viewer');
CREATE TYPE import_source AS ENUM ('ofx', 'csv', 'open_finance', 'whatsapp_bot', 'manual');
CREATE TYPE goal_type AS ENUM ('savings', 'debt_payoff', 'investment', 'emergency_fund', 'custom');
CREATE TYPE recurrence_type AS ENUM ('daily', 'weekly', 'biweekly', 'monthly', 'bimonthly', 'quarterly', 'semiannual', 'annual');
CREATE TYPE ai_suggestion_type AS ENUM ('cut_expense', 'rebalance_portfolio', 'budget_adjust', 'goal_suggestion', 'alert');
CREATE TYPE whatsapp_session_state AS ENUM ('idle', 'awaiting_amount', 'awaiting_category', 'awaiting_description', 'awaiting_account', 'confirming');

-- ============================================================
-- USERS & AUTH
-- ============================================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    full_name VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    phone VARCHAR(20),
    whatsapp_number VARCHAR(20),
    plan plan_type NOT NULL DEFAULT 'free',
    plan_expires_at TIMESTAMPTZ,
    timezone VARCHAR(50) NOT NULL DEFAULT 'America/Sao_Paulo',
    locale VARCHAR(10) NOT NULL DEFAULT 'pt-BR',
    currency VARCHAR(3) NOT NULL DEFAULT 'BRL',
    onboarding_completed BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    email_verified_at TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    device_info JSONB,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE email_verifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE password_resets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- FAMILY / GROUPS
-- ============================================================

CREATE TABLE family_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id),
    invite_code VARCHAR(20) UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE family_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    family_group_id UUID NOT NULL REFERENCES family_groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role family_role NOT NULL DEFAULT 'member',
    can_view_transactions BOOLEAN NOT NULL DEFAULT TRUE,
    can_view_investments BOOLEAN NOT NULL DEFAULT FALSE,
    can_add_transactions BOOLEAN NOT NULL DEFAULT TRUE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(family_group_id, user_id)
);

-- ============================================================
-- CATEGORIES
-- ============================================================

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE, -- NULL = sistema
    family_group_id UUID REFERENCES family_groups(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    icon VARCHAR(50),
    color VARCHAR(7),
    type transaction_type NOT NULL,
    parent_id UUID REFERENCES categories(id),
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Categorias padrão do sistema serão inseridas via seed
-- Ex: Alimentação, Transporte, Saúde, Lazer, Salário, etc.

-- ============================================================
-- ACCOUNTS (Contas bancárias, carteiras, cartões)
-- ============================================================

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    family_group_id UUID REFERENCES family_groups(id),
    name VARCHAR(100) NOT NULL,
    type account_type NOT NULL,
    institution_name VARCHAR(100),
    institution_logo_url TEXT,
    color VARCHAR(7),
    icon VARCHAR(50),
    currency VARCHAR(3) NOT NULL DEFAULT 'BRL',
    initial_balance DECIMAL(15,2) NOT NULL DEFAULT 0,
    current_balance DECIMAL(15,2) NOT NULL DEFAULT 0,
    credit_limit DECIMAL(15,2), -- para cartão de crédito
    credit_closing_day INTEGER, -- dia de fechamento
    credit_due_day INTEGER, -- dia de vencimento
    open_finance_account_id VARCHAR(255), -- integração Open Finance
    open_finance_bank_id VARCHAR(100),
    is_shared BOOLEAN NOT NULL DEFAULT FALSE, -- visível para família
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    include_in_net_worth BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    last_sync_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- TRANSACTIONS
-- ============================================================

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id),
    category_id UUID REFERENCES categories(id),
    family_group_id UUID REFERENCES family_groups(id),
    type transaction_type NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    description VARCHAR(255),
    notes TEXT,
    date DATE NOT NULL,
    is_reconciled BOOLEAN NOT NULL DEFAULT FALSE,
    is_ignored BOOLEAN NOT NULL DEFAULT FALSE, -- ignorar no orçamento
    import_source import_source,
    import_id VARCHAR(255), -- id externo para evitar duplicatas
    transfer_pair_id UUID REFERENCES transactions(id), -- para transferências
    recurrence_id UUID, -- FK para recurrences (abaixo)
    tags TEXT[], -- array de tags livres
    attachments JSONB, -- [{name, url, type}]
    ai_categorized BOOLEAN NOT NULL DEFAULT FALSE,
    ai_confidence DECIMAL(4,3), -- 0.000 a 1.000
    whatsapp_message_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_user_date ON transactions(user_id, date DESC);
CREATE INDEX idx_transactions_account ON transactions(account_id);
CREATE INDEX idx_transactions_category ON transactions(category_id);
CREATE INDEX idx_transactions_import ON transactions(import_source, import_id);

-- ============================================================
-- RECURRENCES (Transações recorrentes)
-- ============================================================

CREATE TABLE recurrences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id),
    category_id UUID REFERENCES categories(id),
    type transaction_type NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    description VARCHAR(255) NOT NULL,
    recurrence_type recurrence_type NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    next_due_date DATE NOT NULL,
    auto_launch BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE transactions ADD CONSTRAINT fk_transaction_recurrence
    FOREIGN KEY (recurrence_id) REFERENCES recurrences(id);

-- ============================================================
-- BUDGETS (Orçamentos)
-- ============================================================

CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    family_group_id UUID REFERENCES family_groups(id),
    category_id UUID REFERENCES categories(id), -- NULL = orçamento geral
    name VARCHAR(100),
    amount DECIMAL(15,2) NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    alert_threshold DECIMAL(4,3) DEFAULT 0.8, -- alertar em 80%
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_budgets_user_period ON budgets(user_id, period_start, period_end);

-- ============================================================
-- INVESTMENTS — Portfolio
-- ============================================================

CREATE TABLE investment_portfolios (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL DEFAULT 'Minha Carteira',
    description TEXT,
    is_default BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- INVESTMENTS — Assets (Ativos)
-- ============================================================

CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticker VARCHAR(20),
    name VARCHAR(255) NOT NULL,
    asset_class asset_class NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'BRL',
    exchange VARCHAR(20), -- B3, NYSE, NASDAQ, etc
    sector VARCHAR(100),
    description TEXT,
    logo_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    -- Preço atual (atualizado por worker)
    current_price DECIMAL(15,6),
    current_price_updated_at TIMESTAMPTZ,
    -- Metadados extras (dividendos, DY, P/VP, etc)
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(ticker, exchange)
);

CREATE INDEX idx_assets_ticker ON assets(ticker);
CREATE INDEX idx_assets_class ON assets(asset_class);

-- Ativo customizado pelo usuário (imóvel, veículo, etc)
CREATE TABLE custom_assets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    asset_class asset_class NOT NULL DEFAULT 'custom',
    description TEXT,
    monthly_income DECIMAL(15,2) DEFAULT 0, -- aluguel, dividendo manual
    current_value DECIMAL(15,2) NOT NULL,
    purchase_value DECIMAL(15,2),
    purchase_date DATE,
    notes TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- INVESTMENTS — Holdings (Posições)
-- ============================================================

CREATE TABLE holdings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    portfolio_id UUID NOT NULL REFERENCES investment_portfolios(id),
    asset_id UUID REFERENCES assets(id),
    custom_asset_id UUID REFERENCES custom_assets(id),
    broker VARCHAR(100),
    quantity DECIMAL(20,8) NOT NULL DEFAULT 0,
    average_price DECIMAL(15,6) NOT NULL DEFAULT 0,
    total_invested DECIMAL(15,2) NOT NULL DEFAULT 0,
    -- Calculados
    current_value DECIMAL(15,2),
    unrealized_pnl DECIMAL(15,2),
    unrealized_pnl_pct DECIMAL(8,4),
    realized_pnl DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_dividends DECIMAL(15,2) NOT NULL DEFAULT 0,
    last_updated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT check_asset_xor CHECK (
        (asset_id IS NOT NULL AND custom_asset_id IS NULL) OR
        (asset_id IS NULL AND custom_asset_id IS NOT NULL)
    )
);

CREATE INDEX idx_holdings_user ON holdings(user_id);
CREATE INDEX idx_holdings_portfolio ON holdings(portfolio_id);

-- ============================================================
-- INVESTMENTS — Transactions (Ordens / Operações)
-- ============================================================

CREATE TABLE investment_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    holding_id UUID NOT NULL REFERENCES holdings(id),
    type VARCHAR(20) NOT NULL CHECK (type IN ('buy', 'sell', 'dividend', 'jcp', 'income', 'split', 'bonus', 'transfer_in', 'transfer_out')),
    quantity DECIMAL(20,8),
    price DECIMAL(15,6),
    amount DECIMAL(15,2) NOT NULL,
    fees DECIMAL(15,2) NOT NULL DEFAULT 0,
    taxes DECIMAL(15,2) NOT NULL DEFAULT 0,
    ir_withheld DECIMAL(15,2) NOT NULL DEFAULT 0,
    date DATE NOT NULL,
    broker VARCHAR(100),
    notes TEXT,
    import_source import_source,
    import_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inv_transactions_holding ON investment_transactions(holding_id);
CREATE INDEX idx_inv_transactions_date ON investment_transactions(user_id, date DESC);

-- ============================================================
-- INVESTMENTS — Renda Fixa (detalhes específicos)
-- ============================================================

CREATE TABLE fixed_income_details (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    holding_id UUID NOT NULL UNIQUE REFERENCES holdings(id) ON DELETE CASCADE,
    fixed_income_type fixed_income_type NOT NULL,
    indexer indexer_type NOT NULL,
    indexer_rate DECIMAL(8,4), -- % sobre indexador (ex: 110% do CDI)
    fixed_rate DECIMAL(8,4), -- taxa fixa (ex: 5.5% a.a.)
    issue_date DATE NOT NULL,
    maturity_date DATE NOT NULL,
    invested_amount DECIMAL(15,2) NOT NULL,
    gross_amount DECIMAL(15,2), -- calculado na vencimento
    net_amount DECIMAL(15,2), -- após IR
    ir_table JSONB, -- tabela regressiva IR
    is_ir_exempt BOOLEAN NOT NULL DEFAULT FALSE, -- LCI, LCA
    institution VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- PRICE HISTORY (Histórico de preços)
-- ============================================================

CREATE TABLE price_history (
    id BIGSERIAL PRIMARY KEY,
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    open DECIMAL(15,6),
    high DECIMAL(15,6),
    low DECIMAL(15,6),
    close DECIMAL(15,6) NOT NULL,
    volume BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(asset_id, date)
);

CREATE INDEX idx_price_history_asset_date ON price_history(asset_id, date DESC);

-- ============================================================
-- GOALS (Metas financeiras)
-- ============================================================

CREATE TABLE goals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type goal_type NOT NULL,
    target_amount DECIMAL(15,2) NOT NULL,
    current_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    target_date DATE,
    monthly_contribution DECIMAL(15,2),
    icon VARCHAR(50),
    color VARCHAR(7),
    description TEXT,
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- CUSTOM FIELDS (Campos customizados pelo usuário)
-- ============================================================

CREATE TABLE custom_fields (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL, -- 'transaction', 'account', 'asset', etc
    name VARCHAR(100) NOT NULL,
    field_type VARCHAR(20) NOT NULL CHECK (field_type IN ('text', 'number', 'date', 'boolean', 'select')),
    options JSONB, -- para select
    is_required BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE custom_field_values (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    field_id UUID NOT NULL REFERENCES custom_fields(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL,
    value TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- DASHBOARDS (Customizáveis)
-- ============================================================

CREATE TABLE dashboards (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    layout JSONB NOT NULL DEFAULT '[]', -- [{widgetId, x, y, w, h}]
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE dashboard_widgets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dashboard_id UUID NOT NULL REFERENCES dashboards(id) ON DELETE CASCADE,
    widget_type VARCHAR(50) NOT NULL, -- 'balance', 'expenses_chart', 'portfolio', etc
    title VARCHAR(100),
    config JSONB NOT NULL DEFAULT '{}',
    position_x INTEGER NOT NULL DEFAULT 0,
    position_y INTEGER NOT NULL DEFAULT 0,
    width INTEGER NOT NULL DEFAULT 2,
    height INTEGER NOT NULL DEFAULT 2,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- NOTIFICATIONS
-- ============================================================

CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    channels notification_channel[] NOT NULL DEFAULT '{email}',
    budget_alerts BOOLEAN NOT NULL DEFAULT TRUE,
    goal_updates BOOLEAN NOT NULL DEFAULT TRUE,
    investment_updates BOOLEAN NOT NULL DEFAULT TRUE,
    bill_reminders BOOLEAN NOT NULL DEFAULT TRUE,
    weekly_summary BOOLEAN NOT NULL DEFAULT TRUE,
    monthly_report BOOLEAN NOT NULL DEFAULT TRUE,
    whatsapp_bot_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    push_token TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type notification_type NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    channel notification_channel NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    sent_at TIMESTAMPTZ,
    read_at TIMESTAMPTZ,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id, is_read, created_at DESC);

-- ============================================================
-- WHATSAPP BOT
-- ============================================================

CREATE TABLE whatsapp_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    phone_number VARCHAR(20) NOT NULL,
    state whatsapp_session_state NOT NULL DEFAULT 'idle',
    context JSONB NOT NULL DEFAULT '{}', -- dados parciais da transação em construção
    last_message_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(phone_number)
);

CREATE TABLE whatsapp_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES whatsapp_sessions(id),
    external_id VARCHAR(255),
    direction VARCHAR(10) NOT NULL CHECK (direction IN ('in', 'out')),
    content TEXT NOT NULL,
    processed BOOLEAN NOT NULL DEFAULT FALSE,
    transaction_id UUID REFERENCES transactions(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- AI SUGGESTIONS
-- ============================================================

CREATE TABLE ai_suggestions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type ai_suggestion_type NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    data JSONB,
    is_dismissed BOOLEAN NOT NULL DEFAULT FALSE,
    is_applied BOOLEAN NOT NULL DEFAULT FALSE,
    dismissed_at TIMESTAMPTZ,
    applied_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- IMPORTS
-- ============================================================

CREATE TABLE import_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID REFERENCES accounts(id),
    source import_source NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    file_name VARCHAR(255),
    file_url TEXT,
    total_rows INTEGER,
    processed_rows INTEGER NOT NULL DEFAULT 0,
    imported_rows INTEGER NOT NULL DEFAULT 0,
    duplicate_rows INTEGER NOT NULL DEFAULT 0,
    error_rows INTEGER NOT NULL DEFAULT 0,
    errors JSONB,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- OPEN FINANCE
-- ============================================================

CREATE TABLE open_finance_consents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bank_id VARCHAR(100) NOT NULL,
    bank_name VARCHAR(100) NOT NULL,
    consent_id VARCHAR(255) NOT NULL UNIQUE,
    access_token_encrypted TEXT,
    refresh_token_encrypted TEXT,
    token_expires_at TIMESTAMPTZ,
    permissions JSONB,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_sync_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- SUBSCRIPTIONS / PLANS
-- ============================================================

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan plan_type NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'canceled', 'past_due', 'trialing')),
    payment_provider VARCHAR(20), -- 'stripe', 'lemon_squeezy', etc
    external_subscription_id VARCHAR(255),
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    canceled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- AUDIT LOG
-- ============================================================

CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),
    entity_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user ON audit_logs(user_id, created_at DESC);

-- ============================================================
-- FUNCTIONS & TRIGGERS
-- ============================================================

-- Auto-update updated_at
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Aplicar em todas as tabelas com updated_at
DO $$
DECLARE
    t TEXT;
BEGIN
    FOR t IN SELECT table_name FROM information_schema.columns
             WHERE column_name = 'updated_at'
             AND table_schema = 'public'
             AND table_name NOT IN ('price_history')
    LOOP
        EXECUTE format('CREATE TRIGGER trg_%s_updated_at
            BEFORE UPDATE ON %I
            FOR EACH ROW EXECUTE FUNCTION update_updated_at()', t, t);
    END LOOP;
END $$;

-- Recalcular balance da conta após transação
CREATE OR REPLACE FUNCTION recalc_account_balance()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.type = 'income' THEN
            UPDATE accounts SET current_balance = current_balance + NEW.amount WHERE id = NEW.account_id;
        ELSIF NEW.type = 'expense' THEN
            UPDATE accounts SET current_balance = current_balance - NEW.amount WHERE id = NEW.account_id;
        END IF;
    ELSIF TG_OP = 'DELETE' THEN
        IF OLD.type = 'income' THEN
            UPDATE accounts SET current_balance = current_balance - OLD.amount WHERE id = OLD.account_id;
        ELSIF OLD.type = 'expense' THEN
            UPDATE accounts SET current_balance = current_balance + OLD.amount WHERE id = OLD.account_id;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_transaction_balance
    AFTER INSERT OR DELETE ON transactions
    FOR EACH ROW EXECUTE FUNCTION recalc_account_balance();

-- ============================================================
-- SEED: Categorias padrão do sistema
-- ============================================================

INSERT INTO categories (id, user_id, name, icon, color, type, is_system, sort_order) VALUES
-- Despesas
(uuid_generate_v4(), NULL, 'Alimentação', 'fork-knife', '#FF6B6B', 'expense', TRUE, 1),
(uuid_generate_v4(), NULL, 'Transporte', 'car', '#FF9F43', 'expense', TRUE, 2),
(uuid_generate_v4(), NULL, 'Moradia', 'home', '#48DBB4', 'expense', TRUE, 3),
(uuid_generate_v4(), NULL, 'Saúde', 'heart', '#FF6B9D', 'expense', TRUE, 4),
(uuid_generate_v4(), NULL, 'Educação', 'book', '#54A0FF', 'expense', TRUE, 5),
(uuid_generate_v4(), NULL, 'Lazer', 'gamepad', '#9B59B6', 'expense', TRUE, 6),
(uuid_generate_v4(), NULL, 'Vestuário', 'shirt', '#F9CA24', 'expense', TRUE, 7),
(uuid_generate_v4(), NULL, 'Assinaturas', 'credit-card', '#6AB04C', 'expense', TRUE, 8),
(uuid_generate_v4(), NULL, 'Pets', 'paw', '#E17055', 'expense', TRUE, 9),
(uuid_generate_v4(), NULL, 'Impostos', 'file-text', '#636E72', 'expense', TRUE, 10),
(uuid_generate_v4(), NULL, 'Outros', 'more-horizontal', '#B2BEC3', 'expense', TRUE, 99),
-- Receitas
(uuid_generate_v4(), NULL, 'Salário', 'briefcase', '#00B894', 'income', TRUE, 1),
(uuid_generate_v4(), NULL, 'Freelance', 'laptop', '#00CEC9', 'income', TRUE, 2),
(uuid_generate_v4(), NULL, 'Investimentos', 'trending-up', '#6C5CE7', 'income', TRUE, 3),
(uuid_generate_v4(), NULL, 'Aluguel Recebido', 'home', '#FDCB6E', 'income', TRUE, 4),
(uuid_generate_v4(), NULL, 'Outras Receitas', 'plus-circle', '#B2BEC3', 'income', TRUE, 99);