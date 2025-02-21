FROM golang:1.21-alpine

WORKDIR /app

RUN apk add --no-cache gcc musl-dev


COPY go.mod go.sum ./
RUN go mod download


COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -o main .


RUN apk add --no-cache tzdata
ENV TZ=America/Sao_Paulo

EXPOSE 3000

CMD ["./main"]