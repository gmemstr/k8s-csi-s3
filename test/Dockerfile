FROM golang:1.24-bullseye

ARG TARGETARCH

LABEL maintainers="Vitaliy Filippov <vitalif@yourcmc.ru>"
LABEL description="csi-s3 testing image"

RUN git clone --depth=1 https://github.com/minio/minio
RUN cd minio && go build && mv minio /usr/local/bin

WORKDIR /build

# prewarm go mod cache
COPY go.mod .
COPY go.sum .
RUN go mod download

RUN wget https://github.com/yandex-cloud/geesefs/releases/latest/download/geesefs-linux-amd64 \
    -O /usr/bin/geesefs && chmod 755 /usr/bin/geesefs

RUN wget https://github.com/tigrisdata/tigrisfs/releases/latest/download/tigrisfs_1.2.0_linux_$TARGETARCH.deb && \
    dpkg -i tigrisfs_1.2.0_linux_$TARGETARCH.deb

ENTRYPOINT ["/build/test/test.sh"]
