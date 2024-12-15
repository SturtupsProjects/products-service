FROM golang:1.23.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./

COPY .env ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/app/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app .
COPY --from=builder /app/main .

RUN mkdir -p pkg/logs

EXPOSE 8070

RUN chmod +x ./main

CMD ["./main"]
