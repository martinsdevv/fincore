CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    type VARCHAR(10) NOT NULL,
    amount BIGINT NOT NULL,
    description VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    transaction_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Garante que o tipo seja apenas 'income' ou 'expense'
    CONSTRAINT chk_transaction_type CHECK (type IN ('income', 'expense')),
    -- Garante que o valor seja sempre positivo (a lógica de negócio decide se soma ou subtrai)
    CONSTRAINT chk_transaction_amount CHECK (amount > 0)
);

-- Índice para acelerar a busca de transações por conta (extratos)
CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);

-- Índice para acelerar a busca por categoria (relatórios)
CREATE INDEX IF NOT EXISTS idx_transactions_category ON transactions(category);
