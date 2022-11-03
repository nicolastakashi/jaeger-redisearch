# syntax=docker/dockerfile:1

FROM gcr.io/distroless/static:latest-amd64

ADD ./bin/jaeger-redisearch-linux-amd64 /go/bin/jaeger-redisearch

WORKDIR /plugin

USER nobody

# /plugin/ location is defined in jaeger-operator
CMD ["cp", "/go/bin/jaeger-redisearch", "/plugin/jaeger-redisearch"]