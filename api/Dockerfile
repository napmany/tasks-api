# build stage
FROM golang:1.12 as builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

# final stage
FROM scratch
WORKDIR /app
COPY config.yml /app/config.yml
COPY --from=builder /app/api /app/api
EXPOSE 80
ENTRYPOINT ["/app/api"]
