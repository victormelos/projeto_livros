package database

import (
	"database/sql"
	"fmt"
	"log"
	"projeto_livros/config"
	"time"

	_ "github.com/lib/pq"
)

func ConnectDB() (*sql.DB, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar configurações: %v", err)
	}

	var db *sql.DB
	maxRetries := 5
	connectionString := cfg.GetDSN()

	for i := 0; i < maxRetries; i++ {
		log.Printf("Tentativa %d de %d de conexão com o banco de dados...", i+1, maxRetries)

		db, err = sql.Open("postgres", connectionString)
		if err != nil {
			log.Printf("Erro ao abrir conexão: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		// Configurar pool de conexões
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		if err = db.Ping(); err == nil {
			log.Println("Conexão com o banco de dados estabelecida com sucesso!")
			return db, nil
		}

		log.Printf("Erro ao pingar o banco: %v", err)
		db.Close()
		time.Sleep(3 * time.Second)
	}

	return nil, fmt.Errorf("falha ao conectar ao banco após %d tentativas: %v", maxRetries, err)
}
