package middleware

import (
	"log"
	"net/http"
)

// CorsMiddleware handles CORS (Cross-Origin Resource Sharing) for the API
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log para depuração
		log.Printf("CORS: Recebida requisição %s para %s com origem %s",
			r.Method, r.RequestURI, r.Header.Get("Origin"))

		// Permitir solicitações de qualquer origem
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Permitir todos os métodos HTTP necessários para a API
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")

		// Permitir todos os cabeçalhos necessários
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept")

		// Permitir credenciais para cookies
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Definir tempo de cache para preflight (OPTIONS)
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 horas

		// Se for uma requisição OPTIONS (preflight), retornar OK imediatamente
		if r.Method == "OPTIONS" {
			log.Printf("CORS: Respondendo a requisição OPTIONS preflight")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Prosseguir com o próximo handler na cadeia
		next.ServeHTTP(w, r)
	})
}
