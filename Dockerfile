FROM golang:1.13-alpine AS build-env

ARG GITHUB_TOKEN
ARG VERSION

RUN apk --no-cache add \
		git \
		make \
	&& echo "machine github.com login ${GITHUB_TOKEN}" > ~/.netrc

WORKDIR /go/src/github.com/nerocrux/sample-rp

ADD . /go/src/github.com/nerocrux/sample-rp

RUN make rp

FROM alpine:latest

COPY --from=build-env /go/src/github.com/nerocrux/sample-rp/rp /rp
RUN mkdir -p /etc/rp
COPY --from=build-env /go/src/github.com/nerocrux/sample-rp/static /etc/rp/static

RUN apk add --no-cache ca-certificates

RUN addgroup -g 1000 -S app && \
    adduser -u 1000 -S app -G app
USER app

EXPOSE 9001

ENTRYPOINT ["/rp"]
