FROM golang:1.23-alpine AS build

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /calendar ./cmd/calendar

FROM alpine:3.20

COPY --from=build /calendar /calendar

# Настройки приходят из переменных окружения в docker-compose.
ENTRYPOINT ["/calendar"]
