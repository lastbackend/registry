# Build manifest
FROM golang:1.11-alpine3.8 as build

RUN apk add --no-cache ca-certificates

RUN apk add --no-cache \
    linux-headers \
    gcc \
    musl-dev

RUN set -ex \
	&& apk add --no-cache --virtual .build-deps \
    bash \
    git  \
    make \
	\
	&& rm -rf /*.patch

ENV GOPATH /go
ENV GOROOT /usr/local/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

ADD .. $GOPATH/src/github.com/lastbackend/registry
WORKDIR $GOPATH/src/github.com/lastbackend/registry

RUN make build && make install
RUN apk del --purge .build-deps

WORKDIR $GOPATH/bin
RUN rm -rf $GOPATH/pkg \
    && rm -rf $GOPATH/src \
    && rm -rf /var/cache/apk/*


# Production manifest
FROM alpine:3.8 as production

RUN apk add --no-cache ca-certificates

COPY --from=build /usr/bin/registry /usr/bin/registry

EXPOSE 2954 2955