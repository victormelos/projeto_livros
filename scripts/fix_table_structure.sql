-- Verificar se a coluna title existe e removê-la se necessário
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'livros' AND column_name = 'title'
    ) THEN
        ALTER TABLE livros DROP COLUMN title;
    END IF;
END $$;

-- Garantir que a coluna quantity seja do tipo INTEGER
ALTER TABLE livros ALTER COLUMN quantity TYPE INTEGER USING quantity::integer;