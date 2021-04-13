FROM  alpine:latest
LABEL maintainers="ics-dev-team"
LABEL description="IncloudSphere CSI Syncer Driver"

COPY ./syncer /bin/ics-csi-syncer

ENTRYPOINT ["/bin/ics-csi-syncer"]

