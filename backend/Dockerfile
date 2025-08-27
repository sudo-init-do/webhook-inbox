# Build stage: use Go 1.24 (required by go.mod/toolchain)
FROM golang:1.24 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Run stage
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /app/server /app/server
COPY --from=build /app/migrations /app/migrations
ENV SERVER_ADDR=:8080
EXPOSE 8080
ENTRYPOINT ["/app/server"]
