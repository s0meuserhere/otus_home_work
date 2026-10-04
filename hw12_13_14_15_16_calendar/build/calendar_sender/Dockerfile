FROM golang:1.23-alpine AS build

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /calendar_sender ./cmd/calendar_sender

FROM alpine:3.20

COPY --from=build /calendar_sender /calendar_sender

# Настройки приходят из переменных окружения в docker-compose.
ENTRYPOINT ["/calendar_sender"]
