---
title: "Why ClickHouse is the Perfect Backend for High-Throughput Metrics Systems"
date: 2026-02-20T10:00:00Z
slug: "clickhouse-metrics-backend"
tags: ["observability", "metrics", "clickhouse", "software development"]
description: "A compelling case for Clickhouse as a metrics storage solution."
published: true
---

<figure>
  <img src="../../static/images/clickhouse-metrics-post/clickhouse-post-banner-image.png" alt="Clickhouse Metrics Banner Image">
</figure>

# Why ClickHouse is the Perfect Backend for High-Throughput Metrics Systems

If you're building a metrics pipeline that needs to ingest millions of data points per second, query them efficiently, and keep storage costs under control, you've probably evaluated the usual suspects: Prometheus, InfluxDB, TimescaleDB, or even rolling your own time-series solution on top of PostgreSQL.

But there's a compelling case for ClickHouse that often gets overlooked in the metrics backend discussion. After building a production metrics system called Faro that processes 50,000+ metrics per second on commodity hardware, I want to share why ClickHouse's architecture makes it uniquely suited for this workload and why it might be the best choice you haven't considered yet.

## The Metrics Backend Challenge

Before diving into ClickHouse specifics, let's establish what makes metrics backends difficult:

1. **Write-heavy workload**: Applications emit metrics constantly - CPU usage, request latency, error rates, business KPIs. A medium-sized infrastructure might generate 100k+ metrics per second.

2. **Time-series characteristics**: Metrics arrive in temporal order with monotonically increasing timestamps. Data is append-only and almost never updated.

3. **Query patterns**: Most queries aggregate data across time windows ("average CPU over the last hour") filtered by dimensions (host, service, environment).

4. **Storage explosion**: Raw metric data grows quickly. A single host emitting 100 metrics/second generates 8.6M data points per day, 260M per month.

5. **Data lifecycle**: Recent metrics need millisecond precision, but older data can be downsampled. Retention policies vary by use case (7 days for raw, 1 year for hourly aggregates).

The system needs to be fast at writes, efficient at storage, and optimized for analytical queries, all simultaneously. This is where ClickHouse's architecture shines.

## ClickHouse's Columnar Advantage

ClickHouse is a columnar database, which fundamentally changes how data is stored and accessed compared to traditional row-oriented databases.

**Row-oriented storage** (MySQL, PostgreSQL):
```
[timestamp=2024-01-12T10:00:00, metric_name=cpu_usage, value=45.2, host=server-1]
[timestamp=2024-01-12T10:00:01, metric_name=cpu_usage, value=46.1, host=server-1]
[timestamp=2024-01-12T10:00:02, metric_name=cpu_usage, value=44.8, host=server-1]
```

**Columnar storage** (ClickHouse):
```
timestamps:    [2024-01-12T10:00:00, 2024-01-12T10:00:01, 2024-01-12T10:00:02, ...]
metric_names:  [cpu_usage, cpu_usage, cpu_usage, ...]
values:        [45.2, 46.1, 44.8, ...]
hosts:         [server-1, server-1, server-1, ...]
```

When you query "What's the average CPU usage for server-1 in the last hour?", ClickHouse only reads the `timestamps`, `values`, and `hosts` columns. It completely skips reading tags, environment, service, and other columns you're not querying.

For metrics workloads where queries typically aggregate a single metric across time, this means reading 10-20% of the data compared to row-oriented databases.

## Compression: The Secret Weapon

Columnar storage also enables aggressive compression. When values of the same type are stored together, compression algorithms can exploit patterns that don't exist in row-oriented layouts.

Here's the metrics table schema from Faro:

```sql
CREATE TABLE metrics (
    timestamp DateTime64(3) CODEC(Delta, ZSTD),
    metric_name LowCardinality(String),
    value Float64 CODEC(ZSTD),
    tags Map(String, String),
    host LowCardinality(String),
    service LowCardinality(String),
    environment LowCardinality(String)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (metric_name, host, service, timestamp)
TTL timestamp + INTERVAL 7 DAY
SETTINGS index_granularity = 8192;
```

Let's break down the compression strategy:

### Delta + ZSTD on Timestamps

Timestamps are monotonically increasing, so instead of storing:
```
[1705060800000, 1705060801000, 1705060802000, 1705060803000]
```

Delta encoding stores:
```
[1705060800000, +1000, +1000, +1000]
```

This transforms random-looking large integers into a repeating pattern that ZSTD compresses to nearly nothing. In practice, timestamps compress to ~90% smaller size.

### LowCardinality for Dimensions

The `LowCardinality` wrapper tells ClickHouse to use dictionary encoding. If you have 10,000 hosts, ClickHouse stores:

```
Dictionary: ["server-1", "server-2", ..., "server-10000"]
Column data: [0, 0, 0, 1, 1, 1, 2, 2, 2, ...]  // Just indices
```

This provides 8-16x storage savings for columns with fewer than 65,536 unique values. Perfect for hosts, services, environments, and metric names.

### Float64 with ZSTD

Metric values might seem random, but they often have patterns: CPU percentages cluster between 0-100, memory values have similar magnitudes, request counts follow daily patterns. ZSTD exploits these patterns for ~70% compression.

**Real-world result**: The raw metrics table achieves 10:1 compression ratios in production. A terabyte of uncompressed metrics compresses to ~100GB, dramatically reducing storage costs and I/O bandwidth requirements.

## MergeTree: Built for Time-Series

ClickHouse's `MergeTree` engine is explicitly designed for time-series workloads. Understanding its architecture reveals why it's so fast.

### Partitioning by Time

```sql
PARTITION BY toYYYYMM(timestamp)
```

This creates separate physical partitions for each month. When you query "metrics from the last 24 hours," ClickHouse immediately skips partitions from previous months, potentially eliminating 99% of data from consideration before even reading a single row.

Partitioning also enables efficient data lifecycle management:

```sql
TTL timestamp + INTERVAL 7 DAY
```

ClickHouse automatically drops partitions older than 7 days in background merges, no manual cleanup scripts required.

### Primary Key & Sparse Index

```sql
ORDER BY (metric_name, host, service, timestamp)
```

The `ORDER BY` clause defines the primary key and physical sort order. ClickHouse creates a **sparse index** with one entry per 8,192 rows (configurable via `index_granularity`).

Unlike dense indexes (B-trees in MySQL/PostgreSQL) that have an entry for every row, sparse indexes only mark the boundaries of data blocks. This makes the index tiny - a billion-row table might have a 100MB index that fits entirely in memory.

When you query:
```sql
SELECT avg(value)
FROM metrics
WHERE metric_name = 'cpu_usage'
  AND host = 'server-1'
  AND timestamp >= now() - INTERVAL 1 HOUR
```

ClickHouse uses the sparse index to skip entire 8,192-row blocks that don't match your filters. If your metric name is `cpu_usage`, it skips all blocks starting with `api_latency`, `disk_io`, etc.

The key order `(metric_name, host, service, timestamp)` matches the typical query pattern: filter by metric, then by dimensions, then by time range. This is a deliberate design choice based on how metrics are actually queried.

## Materialized Views: Pre-Computed Aggregations

One of ClickHouse's killer features for metrics is materialized views with incremental aggregation. Here's how it works:

```sql
CREATE MATERIALIZED VIEW metrics_1m
ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(minute)
ORDER BY (metric_name, host, service, environment, minute)
AS SELECT
    toStartOfMinute(timestamp) as minute,
    metric_name,
    host,
    service,
    environment,
    avgState(value) as avg_value,
    maxState(value) as max_value,
    minState(value) as min_value,
    sumState(value) as sum_value,
    countState() as count
FROM metrics
GROUP BY minute, metric_name, host, service, environment;
```

This creates a view that automatically computes 1-minute aggregations as data arrives. The magic is in three details:

### 1. AggregatingMergeTree Engine

Unlike `MergeTree`, the `AggregatingMergeTree` engine stores aggregate functions in their **intermediate state**, not final values. The `State` suffix on functions (`avgState`, `maxState`, etc.) means "store the state needed to compute this aggregate incrementally."

When ClickHouse merges data parts in the background, it combines these intermediate states. If you have two 1-minute buckets that get merged:

```
Bucket 1: avg_value = avgState([45, 46, 47])  // state = {sum: 138, count: 3}
Bucket 2: avg_value = avgState([48, 49])      // state = {sum: 97, count: 2}
```

After merge:
```
Combined: avg_value = avgState([45, 46, 47, 48, 49])  // state = {sum: 235, count: 5}
```

This happens automatically in background merges, no cron jobs or manual aggregation pipelines required.

### 2. Querying Aggregated Data

To query the materialized view, use the `Merge` suffix:

```sql
SELECT
    minute,
    avgMerge(avg_value) as cpu_avg,
    maxMerge(max_value) as cpu_max
FROM metrics_1m
WHERE metric_name = 'cpu_usage'
  AND host = 'server-1'
  AND minute >= now() - INTERVAL 1 DAY
ORDER BY minute;
```

`avgMerge` tells ClickHouse to finalize the intermediate state into the actual average. This is blazingly fast because ClickHouse is combining pre-computed aggregates, not scanning raw data.

### 3. Multi-Level Aggregation Hierarchy

Faro implements a three-tier aggregation hierarchy:

- **Raw metrics**: 7-day retention, millisecond precision
- **1-minute aggregates**: 30-day retention
- **1-hour aggregates**: 1-year retention

```sql
CREATE MATERIALIZED VIEW metrics_1h
ENGINE = AggregatingMergeTree()
AS SELECT
    toStartOfHour(minute) as hour,
    metric_name,
    host,
    service,
    environment,
    avgMerge(avg_value) as avg_value,
    maxMerge(max_value) as max_value,
    minMerge(min_value) as min_value,
    sumMerge(sum_value) as sum_value,
    sumMerge(count) as count
FROM metrics_1m
GROUP BY hour, metric_name, host, service, environment;
```

Notice this view queries `metrics_1m`, not `metrics`. We're building hourly aggregates from minute aggregates, creating a pyramid:

```
Raw metrics (7 days) → 1m aggregates (30 days) → 1h aggregates (1 year)
```

**Performance impact**: A Grafana dashboard querying 7 days of data uses the 1-minute view, reading ~10,080 rows instead of ~600M raw metrics. Query latency drops from seconds to milliseconds.

## High-Throughput Ingestion

ClickHouse handles write-heavy workloads elegantly. Here's the ingestion architecture from Faro:

```
Client App → Collector (HTTP) → Kafka → Consumer → ClickHouse
                   ↓              ↓         ↓
               Buffer (1k)   Batching   Bulk Insert
               Flush 10s               (10k rows)
```

### Batching is Critical

ClickHouse is optimized for bulk inserts, not individual row inserts. The consumer service batches metrics:

```csharp
var bulkCopy = new ClickHouseBulkCopy(connection)
{
    DestinationTableName = "metrics",
    BatchSize = 10000,
    MaxDegreeOfParallelism = 4
};

await bulkCopy.WriteToServerAsync(metricBatch);
```

This achieves ~100x better throughput than individual inserts. In our testing, a single ClickHouse node sustained 55,000+ metrics per second on commodity hardware, with theoretical capacity for much higher throughput on optimized infrastructure.

### Kafka as a Buffer

Kafka sits between the collector and ClickHouse for two reasons:

1. **Backpressure handling**: If ClickHouse is temporarily slow (background merge, query spike), Kafka buffers writes without dropping data.

2. **Ordered processing**: Partitioning by `metric_name` ensures all data points for a metric are processed in order, which helps ClickHouse's internal optimizations.

### Idempotent Writes

The Kafka producer is configured with:
```csharp
EnableIdempotence = true,
Acks = All,
MaxInFlight = 5
```

This prevents duplicate metrics if there's a network retry, ensuring exactly-once semantics from client to ClickHouse.

## Real-World Query Performance

Let's look at actual queries from Faro's alerting engine:

```sql
SELECT avg(avg_value)
FROM metrics_1m
WHERE metric_name = 'api.latency'
  AND service = 'checkout-service'
  AND minute >= now() - INTERVAL 5 MINUTE
```

This query evaluates an alert checking if average API latency exceeded a threshold in the last 5 minutes. It runs every 30 seconds.

**Query performance**:
- Scans: ~5 rows (1 per minute)
- Execution time: <10ms
- CPU usage: negligible

Without the `metrics_1m` view, the same query against raw data:
- Scans: ~300k rows (assuming 1k metrics/sec)
- Execution time: ~500ms
- CPU usage: noticeable spike

The materialized view provides a **50x performance improvement** for analytical queries.

## Storage Lifecycle & Cost Efficiency

ClickHouse's TTL feature enables automatic data lifecycle management:

```sql
-- Raw metrics
TTL timestamp + INTERVAL 7 DAY

-- 1-minute aggregates
TTL minute + INTERVAL 30 DAY

-- 1-hour aggregates
TTL hour + INTERVAL 365 DAY
```

During background merges, ClickHouse automatically:
1. Drops expired data parts
2. Merges small parts into larger ones
3. Re-compresses data with updated compression statistics

This means:
- No cron jobs to delete old data
- No manual VACUUM operations
- No partition management scripts
- Storage usage stays predictable

**Cost calculation example**:
- 1M metrics/second = 86.4B metrics/day
- At ~40 bytes per metric after compression = 3.5TB/day raw
- With 7-day retention = 24.5TB storage
- Using S3-backed ClickHouse = ~$600/month storage cost

Compare this to storing uncompressed data (~350TB) at ~$8k/month.

## Direct SQL Querying

Unlike some time-series databases that require learning a custom query language (PromQL, Flux), ClickHouse uses SQL. This has several advantages:

### 1. Existing Tools Work

Grafana has native ClickHouse support. Our dashboards use plain SQL:

```sql
SELECT
  toDateTime(minute) as time,
  avgMerge(avg_value) as value,
  host
FROM metrics_1m
WHERE metric_name = 'cpu_usage'
  AND minute >= $__fromTime AND minute <= $__toTime
  AND host IN ($hosts)
GROUP BY time, host
ORDER BY time
```

The `$__fromTime`, `$__toTime`, and `$hosts` are Grafana variables, no plugins or custom code needed.

### 2. Complex Analytics

SQL's expressiveness handles complex queries easily:

```sql
-- 95th percentile latency by endpoint
SELECT
    extractKeyValuePairs(tags, 'endpoint') as endpoint,
    quantileMerge(0.95)(p95_value) as p95_latency
FROM metrics_1m
WHERE metric_name = 'http.request.duration'
  AND minute >= now() - INTERVAL 1 HOUR
GROUP BY endpoint
ORDER BY p95_latency DESC;
```

### 3. No Query Layer Required

Some architectures put a query service between the database and clients to abstract the data model. With ClickHouse, clients query directly using SQL. This eliminates:
- A microservice to build and maintain
- Serialization/deserialization overhead
- An additional failure point
- Translation between query languages

## Load Test Results

To validate these architectural claims, we ran comprehensive load tests against Faro using k6:

**Light Load Test (5 minutes, 100 VUs):**
- **16.79 million metrics** ingested successfully
- **55,961 metrics/second** sustained throughput
- **75.79ms P95 latency**, 303.93ms P99
- **99.9994% success rate** (1 failure in 167,908 requests)
- **2.9M+ metrics** successfully written to ClickHouse with materialized views

**Latency Breakdown:**
- Median request duration: 18.04ms
- P90: 46.43ms
- P95: 75.79ms
- P99: 303.93ms

**Consumer Performance:**
- Average ClickHouse flush time: 23-32ms per 1,000-metric batch
- Sustained write throughput: 30,000-40,000 metrics/sec to ClickHouse
- Materialized views (1m and 1h aggregations) updating in real-time

These results were achieved on a single macOS development machine with Docker Desktop. Production infrastructure with dedicated resources would significantly exceed these numbers.

## Operational Simplicity

ClickHouse's operational characteristics make it easy to run in production:

### Single Node Simplicity

For many use cases, a single ClickHouse node is sufficient. Our testing on commodity hardware achieved 50,000+ metrics/second, and with production-grade resources (NVMe SSDs, 64GB+ RAM), a single node can handle:
- 100k-1M+ inserts/second (depending on hardware and batch sizes)
- Hundreds of concurrent queries
- Petabytes of compressed data

You can start simple and add replication/sharding later if needed.

### Configuration is Minimal

The core ClickHouse configuration in Faro is just environment variables:

```yaml
CLICKHOUSE_HOST: localhost
CLICKHOUSE_PORT: 8123
CLICKHOUSE_DATABASE: metrics
CLICKHOUSE_USER: metrics_user
CLICKHOUSE_PASSWORD: metrics_pass
```

No tuning of buffer pools, cache sizes, or checkpoint intervals. ClickHouse's defaults are sensible for time-series workloads.

### Monitoring Built-In

ClickHouse exposes system tables for monitoring:

```sql
-- Query performance
SELECT query, query_duration_ms, read_rows, memory_usage
FROM system.query_log
WHERE type = 'QueryFinish'
ORDER BY event_time DESC
LIMIT 10;

-- Storage usage
SELECT
    table,
    formatReadableSize(sum(bytes_on_disk)) as size
FROM system.parts
GROUP BY table;
```

These tables integrate with Grafana for observability without external agents.

## When ClickHouse Makes Sense

ClickHouse isn't a universal solution. Here's when it's the right choice:

**Use ClickHouse when**:
- Write volume is high (>10k events/second)
- Data is append-only or immutable
- Queries are analytical (aggregations, filtering, grouping)
- Storage cost matters
- You need both recent high-resolution data and historical aggregates

**Consider alternatives when**:
- Transactional guarantees are required (ACID transactions)
- Data is updated frequently
- Primary access pattern is key-value lookups
- Write volume is low (<1k events/second)

For metrics, tracing, logs, and events, ClickHouse is usually the optimal choice.

## Conclusion: Architecture Matters

The reason ClickHouse works so well for metrics isn't magic—it's architecture. Columnar storage, aggressive compression, sparse indexing, incremental aggregation, and bulk inserts are all deliberate design choices optimized for analytical workloads.

When you align your problem (metrics ingestion and querying) with a database architected specifically for that workload, you get:
- **10x better compression** than row-oriented databases
- **50-100x faster aggregations** through materialized views
- **50,000+ metrics/second** ingestion on commodity hardware (tested), with capacity to scale higher
- **Automatic data lifecycle** management
- **Operational simplicity** with minimal configuration

If you're building a metrics pipeline, evaluating storage backends, or struggling with the limitations of your current solution, give ClickHouse a serious look. The architecture isn't just theoretically elegant - it delivers measurable, practical advantages for time-series workloads at scale.

---

*The Faro project discussed in this post is open source. The complete implementation, including ClickHouse schemas, consumer code, and Grafana dashboards, demonstrates these concepts in a production-ready metrics system. Detailed load test results showing 55,961 metrics/second sustained throughput are available in `LOAD_TEST_PERFORMANCE_SUMMARY.md`.*
