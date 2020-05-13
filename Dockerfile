FROM golang:1.14 AS builder

WORKDIR /go/src/github.com/Octops/gameserver-events-broadcaster

COPY . .

RUN make build

FROM alpine

WORKDIR /app

COPY --from=builder /go/src/github.com/Octops/gameserver-events-broadcaster/bin/broadcaster /app/

RUN chmod +x broadcaster

ENTRYPOINT ["./broadcaster"]