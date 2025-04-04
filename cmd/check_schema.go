package main

import (
	"database/sql"
	"fmt"
	"log"
	"projeto_livros/config"

	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Erro ao carregar configuração: %v", err)
	}

	db, err := sql.Open("postgres", cfg.GetDSN())
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
