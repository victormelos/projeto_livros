package main

import (
	"log"
	"net/http"
	"projeto_livros/database"
	"projeto_livros/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Conectar ao banco de dados
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal("Não foi possível conectar ao banco após várias tentativas")
	}
	defer db.Close()

	bookHandler := handlers.NewBookHandler(db)
	genreHandler := handlers.NewGenreHandler(db)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Rotas
	r.Route("/books", func(r chi.Router) {
		r.Get("/get-all", bookHandler.GetAllBooks)
		r.Post("/create", bookHandler.CreateBook)
		r.Get("/get", bookHandler.GetBook)
		r.Delete("/delete", bookHandler.DeleteBook)
		r.Put("/update", bookHandler.UpdateBook)
		r.Post("/add-all", bookHandler.CreateAllBooks)
	})

	r.Route("/genres", func(r chi.Router) {
		r.Get("/get-all", genreHandler.GetAllGenres)
		r.Post("/create", genreHandler.CreateGenre)
	})

	// Iniciar servidor
	log.Println("Servidor rodando na porta 3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}
