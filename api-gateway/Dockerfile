FROM golang:1.23.2

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./api-gateway

EXPOSE 8080

CMD ["sh", "-c", "sleep 5 && ./main"]