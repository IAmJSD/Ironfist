FROM golang:alpine AS builder
RUN apk --no-cache add gcc g++ make git
WORKDIR /var/ironfist
COPY . .
RUN go get .
RUN GOOS=linux go build -o app .

FROM alpine
EXPOSE 8000
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /var/ironfist/ .
CMD ["./app"]
