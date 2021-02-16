FROM golang:1.15-alpine AS SERVER

ARG VERSION="nightly"

RUN apk add --update git
RUN mkdir -p src/github.com/nats-io && \
    cd src/github.com/nats-io/ && \
    git clone https://github.com/nats-io/nats-server.git && \
    git clone https://github.com/nats-io/natscli.git

RUN cd src/github.com/nats-io/nats-server && CGO_ENABLED=0 GO111MODULE=off go build -v -a -tags netgo -installsuffix netgo -ldflags "-s -w -X github.com/nats-io/nats-server/server.gitCommit=`git rev-parse --short HEAD`" -o /nats-server
RUN cd src/github.com/nats-io/natscli/nats && go build -ldflags "-s -w -X main.version=${VERSION}}" -o /nats

FROM stedolan/jq:latest AS JQ
FROM synadia/nats-box:latest

COPY --from=JQ /usr/local/bin/jq /usr/local/bin/jq
COPY --from=SERVER /nats-server /nats-server
COPY --from=SERVER /nats /usr/local/bin/nats

COPY README.md /
COPY ngs-server.conf /
COPY entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]

EXPOSE 4222
ENV NATS_URL=jetstream:4222

RUN apk add --update ca-certificates man bash && \
    mkdir -p /usr/share/man/man1 && \
    nats --help-man > /usr/share/man/man1/nats.1
