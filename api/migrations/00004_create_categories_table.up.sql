CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Garante que um usuário não possa ter duas categorias com o mesmo nome
    CONSTRAINT uq_user_category_name UNIQUE (user_id, name),

    -- Chave estrangeira para a tabela de usuários.
    -- Se o usuário for deletado, suas categorias vão junto.
    CONSTRAINT fk_categories_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- (Opcional, mas recomendado)
-- Cria um índice para buscas rápidas de "todas as categorias de um usuário"
CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories(user_id);
