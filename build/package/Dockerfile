# syntax=docker/dockerfile:1

FROM alpine:3.16

ADD ./bin/jaeger-redisearch-linux-amd64 /go/bin/jaeger-redisearch

WORKDIR /plugin

# /plugin/ location is defined in jaeger-operator
CMD ["cp", "/go/bin/jaeger-redisearch", "/plugin/jaeger-redisearch"]