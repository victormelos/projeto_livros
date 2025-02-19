FROM golang:1.21-alpine

WORKDIR /app

# Instalar dependências necessárias
RUN apk add --no-cache gcc musl-dev

# Copiar arquivos de dependência primeiro
COPY go.mod go.sum ./
RUN go mod download

# Copiar o resto do código
COPY . .

# Compilar a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Configurar timezone
RUN apk add --no-cache tzdata
ENV TZ=America/Sao_Paulo

EXPOSE 3000

CMD ["./main"]