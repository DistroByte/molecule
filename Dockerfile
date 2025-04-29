FROM golang:1.23.6 as builder

COPY vendor /go/src/github.com/DistroByte/molecule/vendor
COPY . /go/src/github.com/DistroByte/molecule

WORKDIR /go/src/github.com/DistroByte/molecule

RUN CGO_ENABLED=0 go build -o bin/molecule .

FROM gcr.io/distroless/static-debian12

COPY --from=builder /go/src/github.com/DistroByte/molecule/bin/molecule /app/molecule
COPY --from=builder /go/src/github.com/DistroByte/molecule/web /web

WORKDIR /

ENTRYPOINT ["/app/molecule"]
