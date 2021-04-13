# This build arg allows the specification of a custom base image.
FROM ubuntu:18.04
LABEL maintainer="ics-dev-team"
LABEL description="IncloudSphere CSI Driver"

RUN apt-get update && \
    apt-get install -y e2fsprogs xfsprogs btrfs-progs && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY ./icsphere-csi /bin/ics-csi-driver

ENTRYPOINT ["/bin/ics-csi-driver"]

