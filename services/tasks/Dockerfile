FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /tasks ./services/tasks/cmd/tasks

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /tasks .

EXPOSE 8082

ENV TASKS_PORT=8082
ENV AUTH_MODE=http
ENV AUTH_BASE_URL=http://host.docker.internal:8081

CMD ["./tasks"]
