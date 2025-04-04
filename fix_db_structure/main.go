package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func FixDBStructure() {
	// Parâmetros de conexão com o banco de dados
	dbHost := "localhost"
	dbPort := "5432"
	dbUser := "postgres"
	dbPassword := "postgres"
	dbName := "livros"

	// Permitir sobrescrever com variáveis de ambiente
	if os.Getenv("DB_HOST") != "" {
		dbHost = os.Getenv("DB_HOST")
	}
	if os.Getenv("DB_PORT") != "" {
		dbPort = os.Getenv("DB_PORT")
	}
	if os.Getenv("DB_USER") != "" {
		dbUser = os.Getenv("DB_USER")
	}
	if os.Getenv("DB_PASSWORD") != "" {
		dbPassword = os.Getenv("DB_PASSWORD")
	}
	if os.Getenv("DB_NAME") != "" {
		dbName = os.Getenv("DB_NAME")
	}

	// Criar string DSN
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Conectar ao banco de dados
	db, err := sql.Open("postgres", dsn)
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
