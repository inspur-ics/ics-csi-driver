# This build arg allows the specification of a custom base image.
FROM gcr.io/distroless/static:latest
LABEL maintainers="ics-dev-team"

COPY ./syncer /bin/incloudsphere-syncer

ENTRYPOINT ["/bin/incloudsphere-syncer"]

