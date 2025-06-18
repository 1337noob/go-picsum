FROM golang:1.22.2-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/server /app/server
COPY ./images /app/images
EXPOSE 8080
CMD ["/app/server"]