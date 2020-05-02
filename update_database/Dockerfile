FROM golang:alpine AS builder
RUN apk --no-cache add gcc g++ make git
WORKDIR /var/ironfist
COPY . .
RUN go get .
RUN GOOS=linux go build -o app .

FROM alpine
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /var/ironfist/ .
EXPOSE 7000
CMD ["./app"]
