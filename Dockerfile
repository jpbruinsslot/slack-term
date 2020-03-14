FROM golang:alpine as builder

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	ca-certificates

COPY . /go/src/github.com/erroneousboat/slack-term

RUN set -x \
	&& apk add --no-cache --virtual .build-deps \
		git \
		gcc \
		libc-dev \
		libgcc \
		make \
	&& cd /go/src/github.com/erroneousboat/slack-term \
	&& make build \
	&& mv ./bin/slack-term /usr/bin/slack-term \
	&& apk del .build-deps \
	&& rm -rf /go

FROM alpine:latest

ENV USER root

COPY --from=builder /usr/bin/slack-term /usr/bin/slack-term
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs

ENTRYPOINT stty cols 25 && slack-term -config config
