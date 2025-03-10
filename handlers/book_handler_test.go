package handlers
import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
)
func TestGetAllBooks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Erro ao criar mock do banco de dados: %v", err)
	}
	defer db.Close()
	rows := sqlmock.NewRows([]string{"id", "title", "author", "genre_id"}).
		AddRow(1, "Livro 1", "Autor 1", 1).
		AddRow(2, "Livro 2", "Autor 2", 2)
	mock.ExpectQuery("SELECT (.*)").WillReturnRows(rows)
	bookHandler := NewBookHandler(db)
	req, err := http.NewRequest("GET", "/books/get-all", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(bookHandler.GetAllBooks)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler retornou código de status errado: obteve %v, esperava %v",
			status, http.StatusOK)
	}
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("tipo de conteúdo incorreto: obteve %v, esperava %v",
			contentType, "application/json")
	}
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("resposta não é um JSON válido: %v", err)
		t.Logf("Corpo da resposta: %s", rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectativas não atendidas: %s", err)
	}
}
