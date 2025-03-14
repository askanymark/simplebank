# Build
FROM golang:1.23.0-alpine3.20 AS build
WORKDIR /app
COPY . .

RUN go build -o main main.go

# Run
FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/main .
COPY app.env .
COPY start.sh .
COPY wait-for.sh .
COPY db/migration ./db/migration

EXPOSE 8080 9090
CMD ["/app/main"]
ENTRYPOINT ["/app/start.sh"]