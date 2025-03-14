import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { bookService } from '../services/api';

const BookList = () => {
  const [books, setBooks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  useEffect(() => {
    fetchBooks();
  }, [currentPage]);

  const fetchBooks = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await bookService.getAllBooks(currentPage, 20);
      
      // Tratamento unificado para diferentes formatos de resposta
      if (response.data) {
        // Verificar se os dados estão aninhados em um campo "data"
        const booksData = Array.isArray(response.data.data) ? response.data.data : 
                        Array.isArray(response.data) ? response.data : [];
        
        setBooks(booksData);
        
        // Verificar se existe informação de paginação
        if (response.data.pagination) {
          setTotalPages(response.data.pagination.total_pages || 1);
        } else if (response.data.total_pages) {
          setTotalPages(response.data.total_pages);
        } else {
          setTotalPages(1);
        }
      } else {
        setBooks([]);
        setTotalPages(1);
      }
      
      setLoading(false);
    } catch (err) {
      console.error('Erro ao buscar livros:', err);
      setError(err.message || 'Falha ao buscar livros. Tente novamente mais tarde.');
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    if (window.confirm('Tem certeza que deseja excluir este livro?')) {
      try {
        await bookService.deleteBook(id);
        // Atualiza a lista de livros
        fetchBooks();
      } catch (err) {
        console.error('Erro ao excluir livro:', err);
        setError(err.message || 'Falha ao excluir o livro. Por favor, tente novamente.');
      }
    }
  };

  if (loading) return (
    <div className="container mt-5">
      <div className="text-center">
        <div className="spinner-border" role="status">
          <span className="visually-hidden">Carregando...</span>
        </div>
      </div>
    </div>
  );

  return (
    <div className="container mt-4">
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h2>Lista de livros</h2>
        <Link to="/books/new" className="btn btn-primary">Adicionar novo livro</Link>
      </div>

      {error && <div className="alert alert-danger">{error}</div>}

      {books.length === 0 ? (
        <div className="alert alert-info">Nenhum livro encontrado. Adicione alguns livros para começar!</div>
      ) : (
        <>
          <div className="table-responsive">
            <table className="table table-striped table-hover">
              <thead>
                <tr>
                  <th>Título</th>
                  <th>Autor</th>
                  <th>Quantidade</th>
                  <th>Ações</th>
                </tr>
              </thead>
              <tbody>
                {books.map((book) => (
                  <tr key={book.id}>
                    <td>{book.title || book.name}</td>
                    <td>{book.author || 'N/A'}</td>
                    <td>{book.quantity}</td>
                    <td>
                      <div className="btn-group" role="group">
                        <Link to={`/books/edit/${book.id}`} className="btn btn-sm btn-outline-primary">Editar</Link>
                        <button 
                          onClick={() => handleDelete(book.id)} 
                          className="btn btn-sm btn-outline-danger ms-1"
                        >
                          Excluir
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Controles de paginação */}
          {totalPages > 1 && (
            <nav aria-label="Navegação de páginas" className="mt-4">
              <ul className="pagination justify-content-center">
                <li className={`page-item ${currentPage === 1 ? 'disabled' : ''}`}>
                  <button 
                    className="page-link" 
                    onClick={() => setCurrentPage(prev => Math.max(prev - 1, 1))}
                    disabled={currentPage === 1}
                  >
                    Anterior
                  </button>
                </li>
                
                {Array.from({ length: totalPages }, (_, i) => i + 1).map(page => (
                  <li key={page} className={`page-item ${currentPage === page ? 'active' : ''}`}>
                    <button 
                      className="page-link" 
                      onClick={() => setCurrentPage(page)}
                    >
                      {page}
                    </button>
                  </li>
                ))}
                
                <li className={`page-item ${currentPage === totalPages ? 'disabled' : ''}`}>
                  <button 
                    className="page-link" 
                    onClick={() => setCurrentPage(prev => Math.min(prev + 1, totalPages))}
                    disabled={currentPage === totalPages}
                  >
                    Próximo
                  </button>
                </li>
              </ul>
            </nav>
          )}
        </>
      )}
    </div>
  );
};

export default BookList;