CREATE TABLE IF NOT EXISTS livros (
    id VARCHAR(27) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0,
    author VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS genres (
    id VARCHAR(27) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE livros ADD COLUMN IF NOT EXISTS genre_id VARCHAR(27) REFERENCES genres(id);

CREATE INDEX IF NOT EXISTS idx_livros_name ON livros(name);
CREATE INDEX IF NOT EXISTS idx_genres_name ON genres(name);

INSERT INTO genres (id, name, description) VALUES
    (gen_random_uuid(), 'Romance', 'Obras que focam em relacionamentos e emoções'),
    (gen_random_uuid(), 'Ficção Científica', 'Histórias que envolvem avanços científicos e tecnológicos'),
    (gen_random_uuid(), 'Fantasia', 'Mundos mágicos e criaturas fantásticas'),
    (gen_random_uuid(), 'Terror', 'Histórias de suspense e medo'),
    (gen_random_uuid(), 'Drama', 'Narrativas emocionais e conflitos humanos')
ON CONFLICT (name) DO NOTHING;