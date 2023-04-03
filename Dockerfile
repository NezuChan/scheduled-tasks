FROM golang:1.20-alpine as build-stage

WORKDIR /tmp/build

COPY . .

# Build the project
RUN go build cmd/server/main.go

FROM alpine:3

LABEL name "NezukoChan Scheduled Task"
LABEL maintainer "KagChi"

WORKDIR /app

# Install needed deps
RUN apk add --no-cache tini

COPY --from=build-stage /tmp/build/main main

ENTRYPOINT ["tini", "--"]
CMD ["/app/main"]