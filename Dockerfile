FROM golang:1.19-alpine

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY . .

RUN go mod tidy

EXPOSE 8040

RUN go build -o binary

ENTRYPOINT ["/app/binary"]