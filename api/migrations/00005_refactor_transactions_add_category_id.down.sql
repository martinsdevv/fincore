-- 1. Remove o novo índice
DROP INDEX IF EXISTS idx_transactions_category_id;

-- 2. Remove a chave estrangeira
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS fk_transactions_categories;

-- 3. Remove a nova coluna
ALTER TABLE transactions DROP COLUMN category_id;

-- 4. Recria a coluna 'category' original
ALTER TABLE transactions ADD COLUMN category VARCHAR(100) NOT NULL;

-- 5. Recria o índice original
CREATE INDEX IF NOT EXISTS idx_transactions_category ON transactions(category);
