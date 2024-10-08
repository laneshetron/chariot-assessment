FROM golang:1.18 AS build
WORKDIR /go/src/app

# Download dependencies first to improve build time
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -v -o chariot .

FROM alpine:latest
WORKDIR /root/
COPY --from=build /go/src/app/chariot .
EXPOSE 8080
CMD ["./chariot"]
