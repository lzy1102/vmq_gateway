FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY server/vmq_gateway .
EXPOSE 8080
CMD ["./vmq_gateway"]
