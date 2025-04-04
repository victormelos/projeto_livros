FROM golang:1.21-alpine

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

# Copiar e baixar as dependências do Go primeiro para aproveitar o cache do Docker
COPY go.mod go.sum ./
RUN go mod download

# Copiar o código fonte Go
COPY . .

# Criar diretório para os arquivos do frontend
RUN mkdir -p /app/frontend

# Observação: Para incluir o frontend no container, o build do frontend 
# deve ser realizado antes da construção da imagem Docker e os arquivos
# do frontend devem ser copiados para dentro do contexto Docker

# Configurar o timezone
RUN apk add --no-cache tzdata
ENV TZ=America/Sao_Paulo

# Compilar a aplicação Go
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

EXPOSE 3001

CMD ["./main"]