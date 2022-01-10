FROM golang:1.15-alpine as builder
RUN mkdir /build 
ADD . /build/
COPY . /build/
RUN apk --no-cache add ca-certificates
WORKDIR /build 
COPY go.mod go.sum ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/* /app/
WORKDIR /app

EXPOSE 8080
CMD ["./main"]