# Dockerfile for https://github.com/qwzhou/github_releases_exporter
FROM        golang:alpine3.11 AS build
WORKDIR     /go/src
RUN         apk add --update -t build-deps curl libc-dev gcc libgcc git make 
RUN         git clone https://github.com/qwzhou89/github_releases_exporter.git && \
            cd github_releases_exporter &&  \
            make setup && \
            go build -o /usr/local/bin/github_releases_exporter && \
            apk del --purge build-deps && \
            rm -rf /var/cache/apk/* && \
            rm -rf /go

FROM        alpine:3.11
COPY        --from=build /usr/local/bin/github_releases_exporter /usr/local/bin/github_releases_exporter
WORKDIR     /etc/github-releases-exporter
VOLUME      ["/etc/github-releases-exporter"]
EXPOSE      9222
ENTRYPOINT  ["/usr/local/bin/github_releases_exporter"]
