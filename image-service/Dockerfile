FROM golang:1.16.5-alpine3.14 AS builder
WORKDIR /work
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api cmd/service/main.go

FROM alpine:3.14
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /work/api .
CMD ["./api"]  