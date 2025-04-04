package main

import (
	"database/sql"
	"log"

	"github.com/go-delve/delve/pkg/config"
	_ "github.com/lib/pq"
)

func FixDBStructure() {
	// Carregar configuração
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Erro ao carregar configuração: %v", err)
	}

	// Conectar ao banco de dados
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer db.Close()

	// SQL para verificar e corrigir a estrutura da tabela
	sql := `
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
	`

	// Executar o SQL
	_, err = db.Exec(sql)
	if err != nil {
		log.Fatalf("Erro ao executar SQL: %v", err)
	}

	log.Println("Estrutura da tabela verificada e corrigida com sucesso!")
}
