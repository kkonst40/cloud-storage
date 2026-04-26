FROM golang:1.26.2-alpine3.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENBALE=0 go build -o server -tags=viper_bind_struct ./cmd/app/main.go
RUN CGO_ENBALE=0 go build -o migrator -tags=viper_bind_struct ./cmd/migrator/main.go

FROM alpine:3.23

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/migrator .

COPY --from=builder /app/db/migrations ./db/migrations

CMD ["./server"]