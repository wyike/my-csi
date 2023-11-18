FROM alpine

RUN apk add util-linux \
    e2fsprogs

COPY my-csi /usr/local/bin/my-csi

ENTRYPOINT ["/usr/local/bin/my-csi"]