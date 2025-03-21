import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { bookService, genreService } from '../services/api';

const BookForm = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const isEditMode = !!id;
  
  const [book, setBook] = useState({
    title: '',
    author: '',
    quantity: 0,
    genre_id: ''
  });
  
  const [genres, setGenres] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [genreLoading, setGenreLoading] = useState(true);
  const [genreError, setGenreError] = useState(null);

  useEffect(() => {
    // Buscar gêneros
    const fetchGenres = async () => {
      try {
        setGenreLoading(true);
        setGenreError(null);
        const response = await genreService.getAllGenres();
        
        // Tratamento padronizado da resposta
        const genresData = Array.isArray(response.data.data) ? response.data.data : 
                       Array.isArray(response.data) ? response.data : [];
        
        setGenres(genresData);
        setGenreLoading(false);
      } catch (err) {
        console.error('Erro ao buscar gêneros:', err);
        setGenreError(err.message || 'Falha ao carregar gêneros. Por favor, tente novamente.');
        setGenreLoading(false);
      }
    };

    // Buscar dados do livro se estiver no modo de edição
    const fetchBook = async () => {
      if (isEditMode) {
        try {
          setLoading(true);
          const response = await bookService.getBook(id);
          
          if (response.data) {
            // Padronizando para usar title
            const bookData = response.data;
            setBook({
              id: bookData.id,
              title: bookData.title || bookData.name || '',
              author: bookData.author || '',
              quantity: bookData.quantity || 0,
              genre_id: bookData.genre_id || ''
            });
          }
          setLoading(false);
        } catch (err) {
          setError(err.message || 'Falha ao carregar dados do livro. Por favor, tente novamente.');
          setLoading(false);
          console.error('Erro ao buscar livro:', err);
        }
      }
    };

    fetchGenres();
    fetchBook();
  }, [id, isEditMode]);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setBook(prev => ({
      ...prev,
      [name]: name === 'quantity' ? parseInt(value, 20) || 0 : value
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // Validação básica
    if (!book.title.trim()) {
      setError('O título do livro é obrigatório');
      return;
    }
    
    if (!book.author.trim()) {
      setError('O autor do livro é obrigatório');
      return;
    }
    
    try {
      setLoading(true);
      setError(null);
      
      // Dados padronizados para envio
      const bookData = {
        ...book,
        title: book.title.trim(),
        author: book.author.trim(),
        quantity: parseInt(book.quantity, 10) || 0,
        genre_id: book.genre_id || null
      };
      
      console.log('Enviando dados do livro:', bookData);
      
      if (isEditMode) {
        await bookService.updateBook(bookData);
      } else {
        await bookService.createBook(bookData);
      }
      
      setLoading(false);
      navigate('/books');
    } catch (err) {
      console.error('Erro ao salvar livro:', err);
      setError(err.message || 'Falha ao salvar livro. Por favor, verifique os dados e tente novamente.');
      setLoading(false);
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
      <h2>{isEditMode ? 'Editar Livro' : 'Adicionar Novo Livro'}</h2>
      
      {error && <div className="alert alert-danger">{error}</div>}
      {genreError && <div className="alert alert-warning">{genreError}</div>}
      
      <form onSubmit={handleSubmit}>
        <div className="mb-3">
          <label htmlFor="title" className="form-label">Título</label>
          <input
            type="text"
            className="form-control"
            id="title"
            name="title"
            value={book.title}
            onChange={handleChange}
            required
          />
        </div>
        
        <div className="mb-3">
          <label htmlFor="author" className="form-label">Autor</label>
          <input
            type="text"
            className="form-control"
            id="author"
            name="author"
            value={book.author}
            onChange={handleChange}
            required
          />
        </div>
        
        <div className="mb-3">
          <label htmlFor="quantity" className="form-label">Quantidade</label>
          <input
            type="number"
            className="form-control"
            id="quantity"
            name="quantity"
            min="0"
            value={book.quantity}
            onChange={handleChange}
            required
          />
        </div>
        
        <div className="mb-3">
          <label htmlFor="genre_id" className="form-label">Gênero</label>
          {genreLoading ? (
            <div className="d-flex align-items-center">
              <div className="spinner-border spinner-border-sm me-2" role="status">
                <span className="visually-hidden">Carregando gêneros...</span>
              </div>
              <span>Carregando gêneros...</span>
            </div>
          ) : (
            <select
              className="form-select"
              id="genre_id"
              name="genre_id"
              value={book.genre_id || ''}
              onChange={handleChange}
            >
              <option value="">Selecione um gênero (opcional)</option>
              {genres.map(genre => (
                <option key={genre.id} value={genre.id}>
                  {genre.name}
                </option>
              ))}
            </select>
          )}
        </div>
        
        <div className="d-flex gap-2">
          <button type="submit" className="btn btn-primary" disabled={loading}>
            {loading ? 'Salvando...' : 'Salvar Livro'}
          </button>
          <button 
            type="button" 
            className="btn btn-secondary"
            onClick={() => navigate('/books')}
          >
            Cancelar
          </button>
        </div>
      </form>
    </div>
  );
};

export default BookForm;