---
title: Store
type: docs
menu: components
---

# Store

The store component of Thanos implements the Store API on top of historical data in an object storage bucket. It acts primarily as an API gateway and therefore does not need significant amounts of local disk space. It joins a Thanos cluster on startup and advertises the data it can access.
It keeps a small amount of information about all remote blocks on local disk and keeps it in sync with the bucket. This data is generally safe to delete across restarts at the cost of increased startup times.

```bash
$ thanos store \
    --data-dir        "/local/state/data/dir" \
    --objstore.config-file "bucket.yml"
```

The content of `bucket.yml`:

```yaml
type: GCS
config:
  bucket: example-bucket
```

In general about 1MB of local disk space is required per TSDB block stored in the object storage bucket.

## Flags

[embedmd]:# (flags/store.txt $)
```$
usage: thanos store [<flags>]

store node giving access to blocks in a bucket provider. Now supported GCS, S3,
Azure, Swift and Tencent COS.

Flags:
  -h, --help                     Show context-sensitive help (also try
                                 --help-long and --help-man).
      --version                  Show application version.
      --log.level=info           Log filtering level.
      --log.format=logfmt        Log format to use.
      --tracing.config-file=<file-path>
                                 Path to YAML file with tracing configuration.
                                 See format details:
                                 https://thanos.io/tracing.md/#configuration
      --tracing.config=<content>
                                 Alternative to 'tracing.config-file' flag
                                 (lower priority). Content of YAML file with
                                 tracing configuration. See format details:
                                 https://thanos.io/tracing.md/#configuration
      --http-address="0.0.0.0:10902"
                                 Listen host:port for HTTP endpoints.
      --http-grace-period=2m     Time to wait after an interrupt received for
                                 HTTP Server.
      --grpc-address="0.0.0.0:10901"
                                 Listen ip:port address for gRPC endpoints
                                 (StoreAPI). Make sure this address is routable
                                 from other components.
      --grpc-grace-period=2m     Time to wait after an interrupt received for
                                 GRPC Server.
      --grpc-server-tls-cert=""  TLS Certificate for gRPC server, leave blank to
                                 disable TLS
      --grpc-server-tls-key=""   TLS Key for the gRPC server, leave blank to
                                 disable TLS
      --grpc-server-tls-client-ca=""
                                 TLS CA to verify clients against. If no client
                                 CA is specified, there is no client
                                 verification on server side. (tls.NoClientCert)
      --data-dir="./data"        Data directory in which to cache remote blocks.
      --index-cache-size=250MB   Maximum size of items held in the in-memory
                                 index cache. Ignored if --index-cache.config or
                                 --index-cache.config-file option is specified.
      --index-cache.config-file=<file-path>
                                 Path to YAML file that contains index cache
                                 configuration. See format details:
                                 https://thanos.io/components/store.md/#index-cache
      --index-cache.config=<content>
                                 Alternative to 'index-cache.config-file' flag
                                 (lower priority). Content of YAML file that
                                 contains index cache configuration. See format
                                 details:
                                 https://thanos.io/components/store.md/#index-cache
      --chunk-pool-size=2GB      Maximum size of concurrently allocatable bytes
                                 for chunks.
      --store.grpc.series-sample-limit=0
                                 Maximum amount of samples returned via a single
                                 Series call. 0 means no limit. NOTE: For
                                 efficiency we take 120 as the number of samples
                                 in chunk (it cannot be bigger than that), so
                                 the actual number of samples might be lower,
                                 even though the maximum could be hit.
      --store.grpc.series-max-concurrency=20
                                 Maximum number of concurrent Series calls.
      --objstore.config-file=<file-path>
                                 Path to YAML file that contains object store
                                 configuration. See format details:
                                 https://thanos.io/storage.md/#configuration
      --objstore.config=<content>
                                 Alternative to 'objstore.config-file' flag
                                 (lower priority). Content of YAML file that
                                 contains object store configuration. See format
                                 details:
                                 https://thanos.io/storage.md/#configuration
      --sync-block-duration=3m   Repeat interval for syncing the blocks between
                                 local and remote view.
      --block-sync-concurrency=20
                                 Number of goroutines to use when constructing
                                 index-cache.json blocks from object storage.
      --min-time=0000-01-01T00:00:00Z
                                 Start of time range limit to serve. Thanos
                                 Store will serve only metrics, which happened
                                 later than this value. Option can be a constant
                                 time in RFC3339 format or time duration
                                 relative to current time, such as -1d or 2h45m.
                                 Valid duration units are ms, s, m, h, d, w, y.
      --max-time=9999-12-31T23:59:59Z
                                 End of time range limit to serve. Thanos Store
                                 will serve only blocks, which happened eariler
                                 than this value. Option can be a constant time
                                 in RFC3339 format or time duration relative to
                                 current time, such as -1d or 2h45m. Valid
                                 duration units are ms, s, m, h, d, w, y.
      --selector.relabel-config-file=<file-path>
                                 Path to YAML file that contains relabeling
                                 configuration that allows selecting blocks. It
                                 follows native Prometheus relabel-config
                                 syntax. See format details:
                                 https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config
      --selector.relabel-config=<content>
                                 Alternative to 'selector.relabel-config-file'
                                 flag (lower priority). Content of YAML file
                                 that contains relabeling configuration that
                                 allows selecting blocks. It follows native
                                 Prometheus relabel-config syntax. See format
                                 details:
                                 https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config

```

## Time based partitioning

By default Thanos Store Gateway looks at all the data in Object Store and returns it based on query's time range.

Thanos Store `--min-time`, `--max-time` flags allows you to shard Thanos Store based on constant time or time duration relative to current time.

For example setting: `--min-time=-6w` & `--max-time==-2w` will make Thanos Store Gateway return metrics that fall within `now - 6 weeks` up to `now - 2 weeks` time range.

Constant time needs to be set in RFC3339 format. For example `--min-time=2018-01-01T00:00:00Z`, `--max-time=2019-01-01T23:59:59Z`.

Thanos Store Gateway might not get new blocks immediately, as Time partitioning is partly done in asynchronous block synchronization job, which is by default done every 3 minutes. Additionally some of the Object Store implementations provide eventual read-after-write consistency, which means that Thanos Store might not immediately get newly created & uploaded blocks anyway.

We recommend having overlapping time ranges with Thanos Sidecar and other Thanos Store gateways as this will improve your resiliency to failures.

Thanos Querier deals with overlapping time series by merging them together.

Filtering is done on a Chunk level, so Thanos Store might still return Samples which are outside of `--min-time` & `--max-time`.

## Probes

- Thanos Store exposes two endpoints for probing.
  - `/-/healthy` starts as soon as initial setup completed.
  - `/-/ready` starts after all the bootstrapping completed (e.g initial index building) and ready to serve traffic.

> NOTE: Metric endpoint starts immediately so, make sure you set up readiness probe on designated HTTP `/-/ready` path.

## Index cache

Thanos Store Gateway supports an index cache to speed up postings and series lookups from TSDB blocks indexes. Two types of caches are supported:

- `in-memory` (_default_)
- `memcached`

### In-memory index cache

The `in-memory` index cache is enabled by default and its max size can be configured through the flag `--index-cache-size`.

Alternatively, the `in-memory` index cache can also by configured using `--index-cache.config-file` to reference to the configuration file or `--index-cache.config` to put yaml config directly:

[embedmd]:# (../flags/config_index_cache_in_memory.txt yaml)
```yaml
type: IN-MEMORY
config:
  max_size: 0
  max_item_size: 0
```

All the settings are **optional**:

- `max_size`: overall maximum number of bytes cache can contain. The value should be specified with a bytes unit (ie. `250MB`).
- `max_item_size`: maximum size of single item, in bytes. The value should be specified with a bytes unit (ie. `125MB`).

### Memcached index cache

The `memcached` index cache allows to use [Memcached](https://memcached.org) as cache backend. This cache type is configured using `--index-cache.config-file` to reference to the configuration file or `--index-cache.config` to put yaml config directly:

[embedmd]:# (../flags/config_index_cache_memcached.txt yaml)
```yaml
type: MEMCACHED
config:
  addresses: []
  timeout: 0s
  max_idle_connections: 0
  max_async_concurrency: 0
  max_async_buffer_size: 0
  max_get_multi_concurrency: 0
  max_get_multi_batch_size: 0
  dns_provider_update_interval: 0s
```

The **required** settings are:

- `addresses`: list of memcached addresses, that will get resolved with the [DNS service discovery](../service-discovery.md/#dns-service-discovery) provider.

While the remaining settings are **optional**:

- `timeout`: the socket read/write timeout.
- `max_idle_connections`: maximum number of idle connections that will be maintained per address.
- `max_async_concurrency`: maximum number of concurrent asynchronous operations can occur.
- `max_async_buffer_size`: maximum number of enqueued asynchronous operations allowed.
- `max_get_multi_concurrency`: maximum number of concurrent connections when fetching keys. If set to `0`, the concurrency is unlimited.
- `max_get_multi_batch_size`: maximum number of keys a single underlying operation should fetch. If more keys are specified, internally keys are splitted into multiple batches and fetched concurrently, honoring `max_get_multi_concurrency`. If set to `0`, the batch size is unlimited.
- `dns_provider_update_interval`: the DNS discovery update interval.


## Index Header

In order to query series inside blocks from object storage, Store Gateway has to know certain initial info about each block such as:

* symbols table to unintern string values
* postings offset for posting lookup

In order to achieve so, on startup for each block `index-header` is built from pieces of original block's index and stored on disk.
Such `index-header` file is then mmaped and used by Store Gateway.

### Format (version 1)

The following describes the format of the `index-header` file found in each block store gateway local directory.
It is terminated by a table of contents which serves as an entry point into the index.

```
┌─────────────────────────────┬───────────────────────────────┐
│    magic(0xBAAAD792) <4b>   │      version(1) <1 byte>      │
├─────────────────────────────┬───────────────────────────────┤
│  index version(2) <1 byte>  │ index PostingOffsetTable <8b> │
├─────────────────────────────┴───────────────────────────────┤
│ ┌─────────────────────────────────────────────────────────┐ │
│ │      Symbol Table (exact copy from original index)      │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │      Posting Offset Table (exact copy from index)       │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │                          TOC                            │ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

When the index is written, an arbitrary number of padding bytes may be added between the lined out main sections above. When sequentially scanning through the file, any zero bytes after a section's specified length must be skipped.

Most of the sections described below start with a `len` field. It always specifies the number of bytes just before the trailing CRC32 checksum. The checksum is always calculated over those `len` bytes.

### Symbol Table

See [Symbols](https://github.com/prometheus/prometheus/blob/d782387f814753b0118d402ec8cdbdef01bf9079/tsdb/docs/format/index.md#symbol-table)

### Postings Offset Table

See [Posting Offset Table](https://github.com/prometheus/prometheus/blob/d782387f814753b0118d402ec8cdbdef01bf9079/tsdb/docs/format/index.md#postings-offset-table)

### TOC

The table of contents serves as an entry point to the entire index and points to various sections in the file.
If a reference is zero, it indicates the respective section does not exist and empty results should be returned upon lookup.

```
┌─────────────────────────────────────────┐
│ ref(symbols) <8b>                       │
├─────────────────────────────────────────┤
│ ref(postings offset table) <8b>         │
├─────────────────────────────────────────┤
│ CRC32 <4b>                              │
└─────────────────────────────────────────┘
```
