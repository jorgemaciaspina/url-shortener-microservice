FROM golang:1.26.5 AS builder

WORKDIR /app

COPY go.mod ./
 # Pending go.sum

 COPY . .

 RUN go build -o url-shortener main.go

 FROM debian:bookworm-slim

 WORKDIR /app

 COPY --from=builder /app/url-shortener .

 EXPOSE 8080
 CMD ["./url-shortener"]