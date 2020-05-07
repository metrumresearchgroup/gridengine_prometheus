FROM golang:alpine

WORKDIR /app

COPY . .

RUN go build -o /app/gridengine_exporter cmd/server/main.go

EXPOSE 9081
ENTRYPOINT ["/app/gridengine_exporter"]