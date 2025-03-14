package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"projeto_livros/database"
	"projeto_livros/handlers"
	"projeto_livros/middleware"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal("Não foi possível conectar ao banco após várias tentativas")
	}
	defer db.Close()
	bookHandler := handlers.NewBookHandler(db)
	genreHandler := handlers.NewGenreHandler(db)
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CorsMiddleware)
	// Temporarily comment out the auth middleware for testing
	// r.Use(middleware.AuthMiddleware)
	r.Group(func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})
	})

	// Definição de rotas API
	r.Group(func(r chi.Router) {
		// Rotas de livros mais RESTful
		r.Route("/api/books", func(r chi.Router) {
			// Novos endpoints RESTful
			r.Get("/", bookHandler.GetAllBooks)          // Lista todos os livros
			r.Post("/", bookHandler.CreateBook)          // Cria um livro
			r.Get("/{id}", bookHandler.GetBook)          // Busca um livro pelo ID
			r.Put("/{id}", bookHandler.UpdateBook)       // Atualiza um livro
			r.Delete("/{id}", bookHandler.DeleteBook)    // Remove um livro
			r.Post("/batch", bookHandler.CreateAllBooks) // Cria vários livros
		})

		// Mantém as rotas originais para compatibilidade
		r.Route("/books", func(r chi.Router) {
			r.Get("/get-all", bookHandler.GetAllBooks)
			r.Post("/create", bookHandler.CreateBook)
			r.Get("/get", bookHandler.GetBook)
			r.Delete("/delete", bookHandler.DeleteBook)
			r.Put("/update", bookHandler.UpdateBook)
			r.Post("/add-all", bookHandler.CreateAllBooks)

			// Duplica os endpoints RESTful aqui também
			r.Get("/", bookHandler.GetAllBooks)
			r.Post("/", bookHandler.CreateBook)
			r.Get("/{id}", bookHandler.GetBook)
			r.Put("/{id}", bookHandler.UpdateBook)
			r.Delete("/{id}", bookHandler.DeleteBook)
		})

		// Rotas de gêneros mais RESTful
		r.Route("/api/genres", func(r chi.Router) {
			r.Get("/", genreHandler.GetAllGenres)              // Lista todos os gêneros
			r.Post("/", genreHandler.CreateGenre)              // Cria um gênero
			r.Get("/{id}/books", genreHandler.GetBooksByGenre) // Livros de um gênero
		})

		// Mantém as rotas originais para compatibilidade
		r.Route("/genres", func(r chi.Router) {
			r.Get("/get-all", genreHandler.GetAllGenres)
			r.Post("/create", genreHandler.CreateGenre)
			r.Get("/books", genreHandler.GetBooksByGenre)

			// Duplica os endpoints RESTful aqui também
			r.Get("/", genreHandler.GetAllGenres)
			r.Post("/", genreHandler.CreateGenre)
			r.Get("/{id}/books", genreHandler.GetBooksByGenre)
		})
	})

	// Servir arquivos estáticos do frontend
	workDir, _ := os.Getwd()
	frontendDir := filepath.Join(workDir, "../projeto_livros_frontend/build")

	// Verificar se o diretório existe
	if _, err := os.Stat(frontendDir); os.IsNotExist(err) {
		log.Printf("Aviso: Diretório do frontend não encontrado em %s", frontendDir)
		frontendDir = "./frontend" // Tentar um caminho alternativo
	}

	// Servir arquivos estáticos
	filesDir := http.Dir(frontendDir)
	FileServer(r, "/", filesDir)

	// Redirecionar todas as rotas não-API para o index.html para suportar React Router
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(frontendDir, "index.html"))
	})

	log.Println("Servidor rodando na porta 3001")
	log.Fatal(http.ListenAndServe(":3001", r))
}

// FileServer é uma função adaptada para servir arquivos estáticos
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
