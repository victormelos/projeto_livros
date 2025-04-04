-- Script para adicionar a coluna author à tabela livros
ALTER TABLE livros ADD COLUMN IF NOT EXISTS author VARCHAR(255);

-- Atualizar a coluna para aceitar valores nulos (opcional)
-- Isso já é o comportamento padrão, mas estou explicitando para clareza
ALTER TABLE livros ALTER COLUMN author DROP NOT NULL;

-- Criar um índice para melhorar a performance de buscas por autor
CREATE INDEX IF NOT EXISTS idx_livros_author ON livros(author);

-- Confirmar a alteração
SELECT 
    column_name, 
    data_type, 
    is_nullable
FROM 
    information_schema.columns
WHERE 
    table_name = 'livros' AND column_name = 'author';
