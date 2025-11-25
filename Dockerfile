# Golang build image
FROM golang:1.24-alpine AS build
RUN apk add --no-cache alpine-sdk

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go install github.com/a-h/templ/cmd/templ@latest && \
	templ generate

RUN CGO_ENABLED=1 GOOS=linux go build -o main cmd/api/main.go

# Production image
FROM alpine:3.22.2 AS prod

WORKDIR /app

COPY --from=build /app/main /app/main
COPY --from=build /app/s3 /app/s3
COPY --from=build /app/database /app/database
COPY --from=build /app/favicon.ico /app/favicon.ico

EXPOSE ${PORT}
ENTRYPOINT ["./main"]
