FROM golang:1.16 AS builder

WORKDIR /go/src/github.com/Octops/agones-event-broadcaster

COPY . .

RUN make build

FROM alpine

WORKDIR /app

COPY --from=builder /go/src/github.com/Octops/agones-event-broadcaster/bin/broadcaster /app/

RUN chmod +x broadcaster

ENTRYPOINT ["./broadcaster"]