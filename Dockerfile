FROM golang:1.24.1 as builder

COPY . /go/src/github.com/DistroByte/molocule

WORKDIR /go/src/github.com/DistroByte/molocule

RUN CGO_ENABLED=0 go build -o bin/molocule .

FROM gcr.io/distroless/static-debian12

COPY --from=builder /go/src/github.com/DistroByte/molocule/bin/molocule /bin/molocule
COPY --from=builder /go/src/github.com/DistroByte/molocule/web /web

WORKDIR /

ENTRYPOINT ["/bin/molocule"]