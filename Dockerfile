FROM quay.io/prometheus/golang-builder as builder

ADD .   /go/src/github.com/ambrosus/beat-exporter
WORKDIR /go/src/github.com/ambrosus/beat-exporter

RUN make

#FROM        quay.io/prometheus/busybox:latest
FROM debian

RUN apt update
RUN apt install -y docker

COPY --from=builder /go/src/github.com/ambrosus/beat-exporter/beat-exporter  /bin/beat-exporter

EXPOSE      9479
ENTRYPOINT  [ "/bin/beat-exporter" ]