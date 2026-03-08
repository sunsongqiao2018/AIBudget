FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o /bin/edge-api ./cmd/server

FROM alpine:3.20
COPY --from=build /bin/edge-api /usr/local/bin/edge-api
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/edge-api"]
