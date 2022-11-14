FROM golang as builder

WORKDIR /go/github.com/dcherman/smartoilgauge-exporter

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o smartoilgauge-exporter main.go

FROM gcr.io/distroless/static-debian11

COPY --from=builder /go/github.com/dcherman/smartoilgauge-exporter/smartoilgauge-exporter /bin/smartoilgauge-exporter
ENTRYPOINT [ "/bin/smartoilgauge-exporter" ]