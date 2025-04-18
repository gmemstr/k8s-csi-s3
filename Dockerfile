FROM golang:1.24-alpine as gobuild

ARG TARGETARCH

WORKDIR /build
ADD go.mod go.sum /build/
RUN go mod download -x
ADD cmd /build/cmd
ADD pkg /build/pkg
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o ./s3driver ./cmd/s3driver

FROM alpine:3.21
LABEL maintainers="Vitaliy Filippov <vitalif@yourcmc.ru>"
LABEL description="csi-s3 slim image"

ARG TARGETARCH

RUN apk add --no-cache fuse mailcap rclone
RUN apk add --no-cache -X http://dl-cdn.alpinelinux.org/alpine/edge/community s3fs-fuse

ADD https://github.com/yandex-cloud/geesefs/releases/latest/download/geesefs-linux-$TARGETARCH /usr/bin/geesefs
RUN chmod 755 /usr/bin/geesefs

ADD https://github.com/tigrisdata/tigrisfs/releases/latest/download/tigrisfs_1.2.0_linux_$TARGETARCH.apk /tmp/tigrisfs.apk
RUN apk add --allow-untrusted /tmp/tigrisfs.apk

COPY --from=gobuild /build/s3driver /s3driver
ENTRYPOINT ["/s3driver"]
