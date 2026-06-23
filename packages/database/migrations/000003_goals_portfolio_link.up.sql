-- Link goals to investment portfolios so that portfolio value counts toward goal progress.
ALTER TABLE goals
    ADD COLUMN portfolio_id UUID REFERENCES portfolios(id) ON DELETE SET NULL;
