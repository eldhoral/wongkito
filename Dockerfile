FROM golang:1.19-alpine

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY . .

RUN go mod tidy

EXPOSE 8090

RUN go build -o binary

ENTRYPOINT ["/app/binary"]