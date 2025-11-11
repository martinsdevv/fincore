-- 1. Remove o índice antigo da coluna 'category'
DROP INDEX IF EXISTS idx_transactions_category;

-- 2. Remove a coluna antiga 'category'
ALTER TABLE transactions DROP COLUMN category;

-- 3. Adiciona a nova coluna 'category_id'
-- (Adicionamos como NULLABLE primeiro. Em um cenário real com dados,
-- teríamos que atualizar os valores antes de adicionar o NOT NULL.)
ALTER TABLE transactions ADD COLUMN category_id UUID;

-- 4. Adiciona a chave estrangeira para a tabela categories
-- (ON DELETE SET NULL significa que se uma Categoria for deletada,
-- a transação fica "sem categoria", mas não é deletada.)
ALTER TABLE transactions
    ADD CONSTRAINT fk_transactions_categories
    FOREIGN KEY(category_id) REFERENCES categories(id) ON DELETE SET NULL;

-- 5. Cria um novo índice para a nova coluna 'category_id'
CREATE INDEX IF NOT EXISTS idx_transactions_category_id ON transactions(category_id);
