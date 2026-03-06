FROM golang:1.25 AS builder
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 go build -v -o /usr/local/bin/app ./cmd

FROM gcr.io/distroless/static-debian12
COPY --from=builder /usr/local/bin/app /app
CMD ["/app"]
