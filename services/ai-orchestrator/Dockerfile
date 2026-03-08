FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o /bin/ai-orchestrator ./cmd/server

FROM alpine:3.20
COPY --from=build /bin/ai-orchestrator /usr/local/bin/ai-orchestrator
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/ai-orchestrator"]
