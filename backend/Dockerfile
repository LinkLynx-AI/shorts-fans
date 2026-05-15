FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api \
  && CGO_ENABLED=0 GOOS=linux go build -o /out/worker ./cmd/worker \
  && CGO_ENABLED=0 GOOS=linux go build -o /out/migrate ./cmd/migrate \
  && CGO_ENABLED=0 GOOS=linux go build -o /out/devseed ./cmd/devseed

FROM alpine:3.22

RUN apk add --no-cache curl \
  && addgroup -S app \
  && adduser -S app -G app

WORKDIR /app

COPY --from=builder /out/api /app/api
COPY --from=builder /out/worker /app/worker
COPY --from=builder /out/migrate /app/migrate
COPY --from=builder /out/devseed /app/devseed
COPY db/migrations /app/db/migrations

USER app

EXPOSE 8080

CMD ["/app/api"]
