FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY . ./

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o webhook .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/webhook .

RUN chmod +x /app/webhook

ENTRYPOINT [ "/app/webhook" ]