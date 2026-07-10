FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
RUN CGO_ENABLED=0 go build -o vmq_gateway .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/vmq_gateway .
COPY static/ ./static/
EXPOSE 8080
CMD ["./vmq_gateway"]
