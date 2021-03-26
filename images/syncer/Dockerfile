# This build arg allows the specification of a custom base image.
FROM ubuntu:18.04
LABEL maintainers="ics-dev-team"

COPY ./syncer /bin/incloudsphere-syncer

ENTRYPOINT ["/bin/incloudsphere-syncer"]

