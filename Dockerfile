FROM golang:1.18.3-alpine3.16 AS builder
WORKDIR /app
COPY . .
RUN go build -o main

FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/main .
#copy <directory to copy from host> ./<new Directory in 2nd stage container>
COPY logs ./logs
COPY SSL ./SSL
COPY app.env .
COPY wait-for.sh .


EXPOSE 9091
CMD ["/app/main"]