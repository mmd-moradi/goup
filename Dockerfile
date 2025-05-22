FROM golang:1.23.5-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/api/main.go -o ./docs


RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o goup-api ./cmd/api

FROM alpine:3.19


WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata


COPY --from=builder /app/goup-api .

COPY --from=builder /app/configs ./configs

COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./goup-api"]