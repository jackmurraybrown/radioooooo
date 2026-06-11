FROM golang:1.26-alpine AS builder
WORKDIR /app

# cache dependencies before copying source — only re-downloaded when go.mod changes
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api

FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/api /api

EXPOSE 8080
ENTRYPOINT ["/api"]
