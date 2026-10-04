FROM golang:1.23-alpine

WORKDIR /src
COPY . .

CMD ["go", "test", "-tags=integration", "-count=1", "-v", "./test/integration/..."]
