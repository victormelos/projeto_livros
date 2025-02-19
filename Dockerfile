FROM golang:1.21-alpine

WORKDIR /app

# Instalar dependências de build
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 3000

# Adicionar comando para ver logs
CMD ["./main"]