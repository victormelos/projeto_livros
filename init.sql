CREATE TABLE IF NOT EXISTS livros (
    id VARCHAR(27) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Índice para melhorar a performance das buscas por nome
CREATE INDEX IF NOT EXISTS idx_livros_name ON livros(name); 