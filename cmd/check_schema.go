package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
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

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer db.Close()

	// Consulta para obter a estrutura da tabela livros
	query := `
	SELECT column_name, data_type, character_maximum_length
	FROM information_schema.columns
	WHERE table_name = 'livros'
	ORDER BY ordinal_position;
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Erro ao executar consulta: %v", err)
	}
	defer rows.Close()

	fmt.Println("Estrutura da tabela 'livros':")
	fmt.Println("-----------------------------")
	fmt.Printf("%-20s %-15s %-10s\n", "COLUNA", "TIPO", "TAMANHO")
	fmt.Println("-----------------------------")

	for rows.Next() {
		var columnName, dataType string
		var maxLength sql.NullInt64

		if err := rows.Scan(&columnName, &dataType, &maxLength); err != nil {
			log.Fatalf("Erro ao ler resultado: %v", err)
		}

		lengthStr := "NULL"
		if maxLength.Valid {
			lengthStr = fmt.Sprintf("%d", maxLength.Int64)
		}

		fmt.Printf("%-20s %-15s %-10s\n", columnName, dataType, lengthStr)
	}

	if err = rows.Err(); err != nil {
		log.Fatalf("Erro ao iterar resultados: %v", err)
	}
}
