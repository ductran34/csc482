FROM golang:latest AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY server.go ./

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -o main .

FROM alpine:latest

RUN apk update && \
    apk upgrade && \
    apk add ca-certificates

WORKDIR /

COPY --from=build /app/main ./

# Check results
RUN env && pwd && find .

# Start the application

CMD ["./main"]
