package middleware
import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"github.com/golang-jwt/jwt/v5"
)
type contextKey string
const UserIDKey contextKey = "userID"
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}
func GetUserID(ctx context.Context) string {
	userID, _ := ctx.Value(UserIDKey).(string)
	return userID
}
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")
		if token == "" {
			http.Error(w, "Não autorizado", http.StatusUnauthorized)
			return
		}
		secretKey := os.Getenv("JWT_SECRET")
		if secretKey == "" {
			secretKey = "sua_chave_secreta_para_desenvolvimento" 
		}
		claims := &Claims{}
		parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				http.Error(w, "Assinatura do token inválida", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Token inválido: "+err.Error(), http.StatusUnauthorized)
			return
		}
		if !parsedToken.Valid {
			http.Error(w, "Token inválido", http.StatusUnauthorized)
			return
		}
		if claims.ExpiresAt != nil {
			expirationTime := claims.ExpiresAt.Time
			if time.Now().After(expirationTime) {
				http.Error(w, "Token expirado", http.StatusUnauthorized)
				return
			}
		}
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func GenerateToken(userID string) (string, error) {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "sua_chave_secreta_para_desenvolvimento" 
	}
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "api-livros",
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
