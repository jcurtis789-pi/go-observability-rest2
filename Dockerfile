FROM docker.io/golang:1.21 AS builder

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN go build -o server main.go

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
