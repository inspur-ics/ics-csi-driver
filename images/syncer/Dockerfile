# This build arg allows the specification of a custom base image.
ARG BASE_IMAGE=ubuntu:18.04
FROM ${BASE_IMAGE}
LABEL "maintainer" "ics-dev-team"

COPY ./syncer /bin/incloudsphere-syncer

ENTRYPOINT ["/bin/incloudsphere-syncer"]

