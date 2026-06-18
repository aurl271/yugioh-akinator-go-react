# build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o server ./cmd/server

# runtime stage
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/server ./server

ENV PORT=8080

EXPOSE 8080

CMD ["./server"]
