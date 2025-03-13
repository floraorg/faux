FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o faux .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/faux .
COPY --from=builder /app/views ./views
COPY --from=builder /app/.env ./.env

ENV GIN_MODE=release

EXPOSE ${PORT}

CMD ["sh", "-c", "./faux"] 