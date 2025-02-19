package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	var db *sql.DB
	var err error
	maxRetries := 5

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", "postgres://root:root@db:5432/livros?sslmode=disable")
		if err != nil {
			log.Printf("Tentativa %d: Erro ao abrir conexão: %v", i+1, err)
			time.Sleep(5 * time.Second)
			continue
		}

		err = db.Ping()
		if err == nil {
			log.Println("Conectado com sucesso ao banco de dados!")
			break
		}

		log.Printf("Tentativa %d: Erro ao conectar: %v", i+1, err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		log.Fatal("Não foi possível conectar ao banco após várias tentativas")
	}
	defer db.Close()

	// Criar tabela livros se não existir
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS livros (
			id SERIAL PRIMARY KEY,
			titulo VARCHAR(255) NOT NULL,
			autor VARCHAR(255) NOT NULL,
			ano_publicacao INTEGER
		)
	`)
	if err != nil {
		log.Fatal("Erro ao criar tabela:", err)
	}
	log.Println("Tabela 'livros' verificada/criada com sucesso!")

	// Manter a aplicação rodando
	select {}
}
