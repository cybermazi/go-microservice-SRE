FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

# Stage 2: Use distroless for minimal runtime environment
FROM gcr.io/distroless/base

COPY --from=builder /app/main /main

COPY .env .env

EXPOSE 3002

CMD ["/main"]