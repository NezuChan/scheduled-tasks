FROM golang:1.24-alpine as build-stage

WORKDIR /tmp/build

COPY . .

# Build the project
RUN go build .

FROM alpine:3

LABEL name "NezuChan Scheduled Task"
LABEL maintainer "KagChi"

WORKDIR /app

# Install needed deps
RUN apk add --no-cache tini

COPY --from=build-stage /tmp/build/scheduled-tasks main

ENTRYPOINT ["tini", "--"]
CMD ["/app/main"]