# This build arg allows the specification of a custom base image.
ARG BASE_IMAGE=centos:7
FROM ${BASE_IMAGE}
LABEL "maintainer" "ics-dev-team"

COPY ./icsphere-csi /bin/icsphere-csi

ENTRYPOINT ["/bin/icsphere-csi"]

