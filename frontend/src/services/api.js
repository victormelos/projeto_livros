import axios from 'axios';

// Ensure we're using the correct API URL
const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:3001';

// Cria instância axios com configuração padrão
const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Adiciona interceptor de requisição para depuração
api.interceptors.response.use(
  (response) => {
    console.log(`Resposta de ${response.config.url}:`, response.data);
    return response;
  },
  (error) => {
    const errorMessage = error.response?.data?.error || error.message || 'Erro desconhecido na requisição';
    const errorCode = error.response?.data?.code || error.response?.status;
    console.error(`Erro na API (${errorCode}):`, errorMessage, error.response?.data);
    return Promise.reject({
      message: errorMessage,
      originalError: error,
      status: errorCode || error.response?.status
    });
  }
);


// Métodos da API de livros - padronizados para usar title e author
export const bookService = {
  // Obter todos os livros com paginação
  getAllBooks: (page = 1, perPage = 10) => 
    api.get(`/books`, { params: { page, per_page: perPage } }),
  
  // Obter livro por ID
  getBook: (id) => {
    // Usa o endpoint RESTful
    if (id) {
      return api.get(`/books/${id}`);
    }
    // Fallback para o antigo endpoint para compatibilidade
    return api.get(`/books/get`, { params: { id } });
  },
  
  // Criar um novo livro
  createBook: (bookData) => {
    // Enviando tanto name quanto title para garantir compatibilidade
    const data = {
      ...bookData,
      title: bookData.title || '',
      name: bookData.title || '',
      author: bookData.author || '',
      quantity: bookData.quantity || 0,
      genre_id: bookData.genre_id || null
    };
    
    console.log('Enviando dados para criação de livro:', data);
    return api.post('/books', data);
  },
  
  // Atualizar um livro existente
  updateBook: (bookData) => {
    // Enviando tanto name quanto title para garantir compatibilidade
    const data = {
      ...bookData,
      title: bookData.title || '',
      name: bookData.title || '',
      author: bookData.author || '',
      quantity: bookData.quantity || 0,
      genre_id: bookData.genre_id || null
    };
    
    console.log('Enviando dados para atualização de livro:', data);
    
    // Usa o endpoint RESTful se tiver ID
    if (data.id) {
      return api.put(`/books/${data.id}`, data);
    }
    
    // Fallback para o antigo endpoint para compatibilidade
    return api.put('/books/update', data);
  },
  
  // Excluir um livro
  deleteBook: (id) => {
    // Usa o endpoint RESTful
    if (typeof id === 'string') {
      return api.delete(`/books/${id}`);
    }
    
    // Fallback para o antigo endpoint para compatibilidade
    return api.delete('/books/delete', { data: { id } });
  },
  
  // Criar vários livros de uma vez
  createAllBooks: (booksArray) => 
    api.post('/books/batch', booksArray),
};

export const genreService = {
  // Obter todos os gêneros
  getAllGenres: () => 
    api.get('/genres'),
  
  // Criar um novo gênero
  createGenre: (genreData) => 
    api.post('/genres', genreData),
  
  // Obter livros por gênero
  getBooksByGenre: (genreId) => {
    if (genreId) {
      return api.get(`/genres/${genreId}/books`);
    }
    return api.get(`/genres/books`, { params: { genre_id: genreId } });
  }
};

export default api;