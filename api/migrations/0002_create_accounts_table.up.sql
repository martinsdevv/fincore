CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    balance BIGINT NOT NULL DEFAULT 0, -- Armazenado em centavos
    currency VARCHAR(10) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Criar um índice no user_id para otimizar a listagem de contas por usuário
CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);
