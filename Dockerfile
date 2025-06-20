FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod .
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o exam_bot main/main.go

FROM alpine:latest

RUN apk add --no-cache sqlite-libs

WORKDIR /app

COPY --from=builder /app/exam_bot .
COPY --from=builder /app/prod.yaml .

RUN mkdir -p /app/data

ENV STORAGE_PATH=/app/data/exams.db
ENV LOG_PATH=/app/data/bot.log

CMD ["./exam_bot"]