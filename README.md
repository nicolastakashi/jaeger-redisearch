# Jaeger RediSearch
This is a [Jaeger gRPC storage plugin](https://github.com/jaegertracing/jaeger/tree/main/plugin/storage/grpc) implementation for storing traces in [RediSearch](https://redis.io/docs/stack/search/).

## Project status

This is a community-driven project, and you are welcome to share your issues and feature requests. Pull requests are also greatly appreciated.

## Why RediSearch

RediSearch is a [source-available](https://github.com/RediSearch/RediSearch/blob/master/LICENSE) Redis module that enables querying, secondary indexing, and full-text search for Redis. These features enable multi-field queries, aggregation, exact phrase matching, numeric filtering, and geo filtering for text queries.

## How it works

Jaeger data is stored in 2 tables. The first contains operations encoded in JSON. The second stores' key information about spans for searching and this is indexing spans by duration and tags.

## Build & Run

You can just run the following command, to build your local environment with Jaeger, Redis, Plugin and HotRoad.

```bash
make run-all
```

After this you can access: [Jaeger UI](http://localhost:16686)
 and [HotRoad](http://localhost:8080)