package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	handlers "projeto_livros/internal/delivery/http"
	"projeto_livros/internal/delivery/middleware"
	"projeto_livros/internal/repository/database"
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

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Endpoint direto para atualizar quantidade via query params
	r.Get("/update-quantity", bookHandler.UpdateQuantityDirect)

	// API routes using RESTful conventions
	r.Route("/api/books", func(r chi.Router) {
		r.Get("/", bookHandler.GetAllBooks)                        // Lista todos os livros
		r.Post("/", bookHandler.CreateBook)                        // Cria um livro
		r.Post("/batch", bookHandler.CreateAllBooks)               // Cria vários livros
		r.Get("/{id}", bookHandler.GetBook)                        // Busca um livro pelo ID
		r.Put("/{id}", bookHandler.UpdateBook)                     // Atualiza um livro
		r.Delete("/{id}", bookHandler.DeleteBook)                  // Remove um livro
		r.Post("/update-quantity", bookHandler.UpdateBookQuantity) // Endpoint para atualização de quantidade
	})

	r.Route("/api/genres", func(r chi.Router) {
		r.Get("/", genreHandler.GetAllGenres)              // Lista todos os gêneros
		r.Post("/", genreHandler.CreateGenre)              // Cria um gênero
		r.Get("/{id}/books", genreHandler.GetBooksByGenre) // Livros de um gênero
	})

	// Servir arquivos estáticos do frontend
	workDir, _ := os.Getwd()
	var frontendDir string

	// Verificar os possíveis locais do frontend em ordem de prioridade
	possiblePaths := []string{
		filepath.Join(workDir, "../front-end/build"), // Desenvolvimento local
		"../front-end/build",                         // Relativo à raiz
		"./frontend",                                 // Dentro do container Docker
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			frontendDir = path
			break
		}
	}

	if frontendDir == "" {
		log.Printf("Aviso: Diretório do frontend não encontrado em nenhum local esperado")
		frontendDir = "./frontend" // Fallback padrão
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
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
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
