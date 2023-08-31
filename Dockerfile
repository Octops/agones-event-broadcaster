FROM golang:1.20 AS builder

WORKDIR /go/src/github.com/Octops/agones-event-broadcaster

COPY . .

RUN make build && chmod +x /go/src/github.com/Octops/agones-event-broadcaster/bin/broadcaster

FROM gcr.io/distroless/static:nonroot

WORKDIR /app

COPY --from=builder /go/src/github.com/Octops/agones-event-broadcaster/bin/broadcaster /app/

ENTRYPOINT ["./broadcaster"]
