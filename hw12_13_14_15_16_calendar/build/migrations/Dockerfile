FROM golang:1.23-alpine

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.26.0

COPY migrations /migrations

CMD goose -dir /migrations postgres "host=$DB_HOST port=$DB_PORT user=$DB_USER password=$DB_PASSWORD dbname=$DB_NAME sslmode=disable" up
