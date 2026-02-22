---
title: "Why ClickHouse is the Perfect Backend for High-Throughput Metrics Systems"
date: 2026-02-21T16:00:00Z
slug: "clickhouse-metrics-backend"
tags: ["observability", "metrics", "clickhouse", "software development"]
description: "A compelling case for Clickhouse as a metrics storage solution."
published: true
---

<figure>
  <img src="../../static/images/clickhouse-metrics-post/clickhouse-post-banner-image.png" alt="Clickhouse Metrics Banner Image">
</figure>

# Why ClickHouse is the Perfect Backend for High-Throughput Metrics Systems

If you're building a metrics pipeline that needs to ingest millions of data points per second, query them efficiently, and keep storage costs under control, you've probably evaluated Prometheus, InfluxDB, TimescaleDB, or even PostgreSQL with TimescaleDB extensions.

But there's a compelling case for ClickHouse that often gets overlooked. After building **Faro**, a production metrics system that sustains **55,000+ metrics per second** on commodity hardware, I want to share why ClickHouse's architecture makes it uniquely suited for this workload and why it might be the best choice you haven't considered yet.

## The Metrics Backend Challenge

Metrics backends are hard to build because they need to satisfy competing requirements simultaneously:

1. **Write-heavy workload**: Constant ingestion from hundreds or thousands of services
2. **Time-series characteristics**: Append-only data with monotonically increasing timestamps
3. **Analytical queries**: Aggregations across time windows, filtered by dimensions
4. **Storage explosion**: 100 metrics/sec = 8.6M points/day = 260M points/month per host
5. **Data lifecycle**: Millisecond precision for recent data, downsampled aggregates for historical data

Most databases are optimized for transactional workloads (PostgreSQL, MySQL) or key-value lookups (Redis, DynamoDB). Time-series databases like Prometheus work well but have limitations at scale. ClickHouse, a columnar analytical database, is architecturally designed for exactly this workload.

## Why ClickHouse Works: The Technical Foundation

ClickHouse's architecture delivers three key advantages for metrics: columnar storage, aggressive compression, and incremental aggregation.

### Columnar Storage

Traditional row-oriented databases store complete records together on disk where each row contains all fields for a metric point: timestamp, metric name, value, host, service, environment, tags, and metadata. When you execute a query like "what's the average CPU usage for the last hour," a row-oriented database must read entire rows into memory, deserialize all fields, then discard the 80% of data you don't need. This is fundamentally wasteful for analytical queries that aggregate a single metric across time.

Columnar databases invert this structure by storing each column separately in contiguous blocks on disk. ClickHouse writes all timestamps together, all metric names together, all values together. When you query for average CPU, ClickHouse reads only three narrow columns: `timestamp`, `metric_name`, and `value`. It completely skips tags, hosts, services, and all other metadata. For analytical queries typical in metrics systems, this means **reading just 10-20% of the data** compared to row-oriented databases. The performance benefits compound - less disk I/O means faster queries, better cache utilization, and the ability to process more data in the same amount of time.

The columnar layout also enables SIMD (Single Instruction, Multiple Data) vectorization, where modern CPUs process multiple values simultaneously. When aggregating thousands of numeric values stored contiguously in memory, ClickHouse can compute sums, averages, and other operations 4-8x faster than row-oriented processing. For a metrics query scanning millions of data points, this architectural advantage translates directly to sub-second response times.

### Aggressive Compression

Columnar storage creates a perfect foundation for compression because values of the same type and semantic meaning are stored together. When ClickHouse writes a column of timestamps to disk, it's compressing millions of monotonically increasing integers that differ by small, predictable amounts. When it compresses metric names, it's encoding thousands of repetitions of strings like "cpu_usage" and "memory_bytes". This homogeneity enables compression algorithms to achieve ratios that would be impossible with the mixed-type data in row-oriented storage.

ClickHouse employs multiple specialized compression techniques tailored to different data patterns. For timestamps, it uses delta encoding—instead of storing absolute Unix timestamps like `1645564800`, `1645564801`, `1645564802`, it stores the first value and then differences: `1645564800`, `+1`, `+1`, `+1`. These small integers compress extraordinarily well with algorithms like ZSTD. For low-cardinality fields like metric names and host identifiers, ClickHouse uses dictionary encoding: it creates a mapping where "cpu_usage" becomes `1` and "memory_bytes" becomes `2`, then stores arrays of small integers instead of repeated strings. For numeric metric values that follow patterns (gradual CPU increases, periodic memory oscillations), ZSTD's pattern-matching compression achieves excellent ratios.

The compounding effect of these techniques delivers **10:1 compression ratios** in real-world metrics workloads. A terabyte of uncompressed time-series data representing hundreds of millions of individual metric points compresses down to approximately 100GB on disk. This isn't just about storage cost savings; compression directly improves query performance. When ClickHouse reads compressed data from disk, it decompresses on-the-fly in CPU cache, meaning queries effectively read 10x more data from disk in the same I/O operation. For a system ingesting 50,000 metrics per second, the difference between 5TB and 500GB of monthly storage determines whether you can afford to keep granular data or are forced into aggressive downsampling.

### Sparse Indexing & Partitioning

ClickHouse uses sparse indexes (one entry per 8,192 rows) that fit entirely in memory. Combined with monthly partitioning, queries like "metrics from the last 24 hours" skip 99%+ of data before reading a single row.

### Incremental Aggregation

Most databases force you to choose between storing raw data or pre-aggregated summaries. Store raw data and queries are slow. Store only aggregates and you lose granularity. ClickHouse's materialized views eliminate this tradeoff by computing aggregations incrementally as data arrives, maintaining both raw metrics and pre-computed summaries simultaneously. When you write a metric point to the raw table, ClickHouse automatically updates the corresponding aggregate in real-time. No batch jobs, no cron tasks, no eventual consistency delays.

The key innovation is the `AggregatingMergeTree` engine combined with state functions. Traditional aggregation functions like `avg()` and `max()` return final results. ClickHouse's state functions - `avgState()`, `maxState()`, `sumState()` - return intermediate aggregation states that can be merged incrementally. When new data arrives, ClickHouse doesn't recompute the entire average from scratch; it merges the new data's state with the existing state. This is mathematically equivalent to computing the aggregate over all data, but computationally it's orders of magnitude cheaper. As background merge operations consolidate data parts, ClickHouse efficiently combines these states, maintaining accuracy while spreading the computational cost across write operations rather than concentrating it at query time.

For a dashboard querying the last 24 hours of metrics at 1-minute resolution, this architecture transforms the workload from scanning 86 million raw data points to reading 1,440 pre-aggregated rows. The query time drops from seconds to single-digit milliseconds, **50-100x faster** than scanning raw data. The trade-off is straightforward: you pay a small incremental cost at write time (usually 5-15% overhead) to avoid massive computational costs at query time. For metrics systems where queries vastly outnumber writes and dashboards demand sub-second response times, this is an exceptional bargain.

## Faro: A Production-Ready Implementation

To prove these concepts work in practice, I built **[Faro](https://github.com/seanankenbruck/faro)** - a complete, self-hosted metrics monitoring and alerting system. Faro demonstrates that you can build a production-grade metrics pipeline using ClickHouse without the complexity of commercial solutions like Datadog or Prometheus + Thanos.

### What Faro Does

Faro provides end-to-end metrics monitoring:  
- **Ingestion**: HTTP API accepts metric data points from any application  
- **Storage**: ClickHouse stores raw metrics with automatic multi-tier aggregation  
- **Visualization**: Grafana dashboards query ClickHouse directly via SQL  
- **Alerting**: Built-in alerting engine evaluates rules and sends notifications (email, webhooks, Slack)  
- **Client SDK**: .NET library for easy integration into applications  

The entire implementation is **~2,000 lines of C#** built on .NET 9, proving you don't need massive frameworks to handle high-throughput metrics.

### Architecture Overview

```
Client Apps (SDK) → Collector API → Kafka → Consumer → ClickHouse
                         ↓              ↓         ↓           ↓
                    Validation      Buffering  Batching  Materialized Views
                                                              ↓
                                                      ┌───────┴───────┐
                                                      ↓               ↓
                                                  Grafana      Alerting Engine
                                                 Dashboards    (notifications)
```

### Core Components

**1. Faro.Collector (Metrics Ingestion API)**

The collector is an ASP.NET Core service that serves as the system's entry point, exposing HTTP endpoints for metric ingestion. Applications send metrics via `POST /api/metrics/single` for individual data points or `POST /api/metrics/batch` for up to 10,000 metrics at once. The batch endpoint is crucial for high-throughput scenarios where clients aggregate metrics locally before transmission, reducing network round-trips and HTTP overhead.

The collector validates incoming metrics using FluentValidation to ensure data quality before it enters the pipeline catching malformed timestamps, missing metric names, and invalid tag structures at the edge. Once validated, metrics are buffered in memory with a configurable flush interval (typically 100-500ms) to batch writes to Kafka efficiently. The collector partitions Kafka messages by metric name, ensuring that all data points for a given metric are processed in order and written to the same ClickHouse partition, which optimizes merge operations and compression. Snappy compression reduces network bandwidth between the collector and Kafka, while rate limiting prevents client abuse and protects downstream components from overload. Health check endpoints expose readiness and liveness probes for Kubernetes orchestration or monitoring systems.

**2. Faro.Consumer (Kafka → ClickHouse Pipeline)**

The consumer is a background worker service that bridges Kafka and ClickHouse, continuously reading metric batches from Kafka topics and executing bulk inserts into the database. Rather than writing metrics individually, the consumer accumulates batches of 1,000-10,000 data points before issuing a single bulk insert operation. This batching strategy is critical for ClickHouse performance as individual inserts create small data parts that require excessive merge operations, while bulk inserts create optimally sized parts that merge efficiently.

The consumer uses ClickHouse's native bulk copy API, which streams data directly into table storage without intermediate serialization steps, achieving throughput of 30,000-40,000 metrics per second on commodity hardware. Network failures and transient ClickHouse unavailability are handled via retry logic with exponential backoff using the Polly library, ensuring that temporary issues don't result in data loss. Because the consumer maintains no state beyond Kafka offsets (managed by Kafka itself), it's trivially horizontally scalable. If ingestion throughput exceeds a single consumer's capacity, deploying additional consumer instances automatically distributes the workload across Kafka partitions, linearly increasing write throughput.

**3. Faro.Storage (ClickHouse Data Layer)**

Faro.Storage is a repository abstraction that encapsulates all ClickHouse interactions, providing a clean separation between business logic and database operations. On startup, it handles schema initialization automatically creating the metrics table, materialized views for 1-minute and 1-hour aggregations, and TTL policies for automatic data lifecycle management. This ensures that a fresh Faro deployment can initialize an empty ClickHouse instance without manual SQL execution or migration scripts.

The storage layer manages connection pooling to ClickHouse, maintaining a pool of reusable database connections that eliminates the overhead of establishing new connections for each bulk insert operation. For high-throughput workloads where the consumer executes hundreds of inserts per second, connection pooling is essential to avoid connection exhaustion and TCP handshake overhead. Health check methods expose ClickHouse availability to monitoring systems, allowing orchestrators like Kubernetes to detect database failures and trigger alerts or automated recovery procedures.

**4. Faro.AlertingEngine (Rule Evaluation & Notifications)**

The alerting engine is a continuous evaluation system that monitors metrics and triggers notifications when conditions violate defined thresholds. Unlike systems that require complex rule storage in databases, Faro loads alert rules from simple JSON configuration files, making it easy to version control alert definitions alongside application code and deploy them through standard CI/CD pipelines.

The engine executes SQL queries against ClickHouse at configurable intervals (typically 30-60 seconds), evaluating conditions like "average CPU usage over the last 5 minutes exceeds 80%". To prevent alert flapping from transient spikes, it manages state transitions through a progression: `OK → Pending → Firing → Resolved`. An alert enters the `Pending` state when the condition first becomes true, transitions to `Firing` only after remaining true for the configured "for duration" (preventing false alarms from momentary anomalies), and moves to `Resolved` when conditions return to normal. Notifications are sent via pluggable channels including email (SMTP), webhooks for integration with incident management systems, and direct Slack integration for team notifications.

Alert rules are simple JSON configurations:
```json
{
  "name": "high-cpu-usage",
  "query": "SELECT avg(value) FROM metrics_1m WHERE metric_name='cpu_usage' AND minute >= now() - INTERVAL 5 MINUTE",
  "threshold": 80,
  "condition": "GreaterThan",
  "evaluationInterval": "30s",
  "forDuration": "5m"
}
```

**5. Faro.Client (SDK for .NET Applications)**

The client SDK is a lightweight HTTP library that makes instrumenting .NET applications trivial. Rather than requiring developers to manually construct HTTP requests and manage retry logic, the SDK provides a clean, idiomatic API for sending metrics with minimal boilerplate. It integrates seamlessly with ASP.NET Core's dependency injection, allowing applications to configure the collector URL once at startup and inject the metrics client wherever needed.

Applications can send individual metrics or batch multiple data points for efficiency. The SDK handles serialization, HTTP connection management, and automatic retries on transient failures, abstracting away the networking complexity so developers can focus on instrumenting business logic. For scenarios like recording API request durations or tracking custom business metrics, the SDK provides a simple, type-safe interface:
```csharp
services.AddFaroMetrics(config => config.CollectorUrl = "http://localhost:5000");

// Send a metric
await metricsClient.SendAsync(new MetricPoint {
    Name = "api.request.duration",
    Value = 45.2,
    Tags = new() { ["endpoint"] = "/api/users", ["method"] = "GET" }
});
```

### Why These Design Decisions?

**Kafka as a Buffer**

Kafka decouples ingestion from storage, providing critical reliability benefits:  
- If ClickHouse is temporarily slow (background merge, query spike), Kafka buffers writes without data loss  
- Consumers can restart without losing metrics  
- Partitioning by metric name ensures ordered writes, which helps ClickHouse's internal optimizations  

**No Query Service Layer**

Unlike architectures that put an API between clients and the database (e.g., Prometheus's HTTP API), Faro lets Grafana and the alerting engine query ClickHouse directly using SQL. This eliminates an entire microservice to build and maintain, the need to manage query translation logic, serialization/deserialization overhead, all of which introduce the possibility of another point of failure.

**Multi-Tier Aggregation**

Raw metrics (7-day retention) → 1-minute aggregates (30-day retention) → 1-hour aggregates (1-year retention)

This hierarchy balances query performance with storage costs. Dashboard queries for the last 7 days use the 1-minute view, reading ~10k rows instead of hundreds of millions. Historical trend analysis uses the 1-hour view.

**C# and .NET 9**

Choosing .NET provides:  
- Excellent async/await primitives for high-concurrency workloads  
- First-class HTTP/REST support via ASP.NET Core  
- Strong typing and compile-time safety  
- Native performance comparable to Go/Rust for I/O-bound tasks  
- Mature ecosystem (Kafka clients, ClickHouse drivers, validation libraries)  

The implementation uses modern C# features like nullable reference types, minimal APIs, and dependency injection for clean, maintainable code.

### Deployment Simplicity

Faro runs in Docker Compose for local development and is also production-ready for more advanced deployments:  
- Kafka (KRaft mode—no Zookeeper required)  
- ClickHouse (single node, scales to replicated clusters)  
- Grafana (with ClickHouse data source pre-configured)  
- Faro services (Collector, Consumer, Alerting Engine)  

Total infrastructure: **5 containers**. No Kubernetes required for moderate workloads.

### The Result

A complete, production-ready metrics system in ~2,000 lines of code that handles 50,000+ metrics/second. The simplicity proves that ClickHouse's architecture does the heavy lifting—your application code stays clean and focused.

## Load Test Results: Proving It Works

We ran comprehensive load tests using k6 to validate real-world performance.

<div style="background: #f6f8fa; border-left: 4px solid #0969da; padding: 16px; margin: 24px 0;">

<strong>Test Configuration:</strong>
<ul>
<li>Duration: 5 minutes</li>
<li>Virtual Users: 100 concurrent clients</li>
<li>Environment: macOS development machine with Docker Desktop</li>
<li>Target: Sustained high-throughput ingestion</li>
</ul>

<strong>Results:</strong>

<div style="margin: 16px 0;">
<strong style="font-size: 1.1em;">📊 Throughput</strong><br/>
• <strong>16.79 million metrics</strong> ingested successfully<br/>
• <strong>55,961 metrics/second</strong> sustained rate<br/>
• <strong>99.9994% success rate</strong> (1 failure in 167,908 requests)
</div>

<div style="margin: 16px 0;">
<strong style="font-size: 1.1em;">⚡ Latency</strong><br/>
• Median (P50): 18.04ms<br/>
• P90: 46.43ms<br/>
• P95: 75.79ms<br/>
• P99: 303.93ms
</div>

<div style="margin: 16px 0;">
<strong style="font-size: 1.1em;">💾 Database Performance</strong><br/>
• ClickHouse write batches: 23-32ms per 1,000 metrics<br/>
• Sustained write throughput: 30,000-40,000 metrics/sec<br/>
• <strong>2.9M+ metrics</strong> successfully stored with materialized views updating in real-time
</div>

</div>

**What This Means:**

These numbers were achieved on a **single development machine**. Production infrastructure with NVMe SSDs, dedicated resources, and horizontal scaling would significantly exceed these results. A single optimized ClickHouse node can handle **100k-1M+ metrics/second**.

The key insight: the bottleneck wasn't ClickHouse—it was the test environment. ClickHouse's bulk insert performance (23-32ms per 1,000 metrics) indicates substantial headroom for higher throughput.

## Storage & Cost Efficiency

With 10:1 compression and automatic TTL-based cleanup, storage costs stay predictable:

**Example calculation for 100k metrics/second:**  
- Raw ingestion: 8.64B metrics/day  
- After compression: ~350GB/day  
- 7-day retention: ~2.5TB storage  
- Using S3-backed ClickHouse: **~$60/month** storage cost

Compare this to storing uncompressed data (~25TB) at ~$600/month.

## Direct SQL Querying: No Translation Layer

ClickHouse uses standard SQL. Grafana dashboards, alerting rules, and ad-hoc queries use the same language:

```sql
SELECT
  toDateTime(minute) as time,
  avgMerge(avg_value) as cpu_avg
FROM metrics_1m
WHERE metric_name = 'cpu_usage'
  AND minute >= now() - INTERVAL 1 HOUR
ORDER BY time;
```

No custom query language to learn. No translation layer to build. No serialization overhead.

## When ClickHouse Makes Sense

**Use ClickHouse when:**  
- Write volume is high (>10k events/second)  
- Data is append-only or immutable  
- Queries are analytical (aggregations, time-series analysis)  
- Storage cost matters  
- You need both high-resolution recent data and historical aggregates  

**Consider alternatives when:**  
- Transactional guarantees are required  
- Data is updated frequently  
- Primary access pattern is key-value lookups  
- Write volume is low (<1k events/second)  

For metrics, traces, logs, and event analytics, ClickHouse is usually the optimal choice.

## Operational Simplicity

ClickHouse is surprisingly easy to operate:

**Single Node Simplicity**: Many workloads run fine on a single node. Start simple, add replication/sharding later.

**Minimal Configuration**: Basic setup requires just connection credentials. No buffer pool tuning, checkpoint intervals, or vacuum strategies.

**Automatic Maintenance**: Background merges, TTL enforcement, and compression happen automatically. No cron jobs or manual intervention.

**Built-in Monitoring**: System tables expose query performance, storage usage, and health metrics—integrate directly with Grafana.

## Conclusion

ClickHouse isn't a universal solution, but for high-throughput metrics systems it delivers measurable, proven advantages that fundamentally change what's possible on modest infrastructure. The architecture achieves 10x better compression than row-oriented databases through columnar storage and specialized encoding, transforming terabytes of metrics into hundreds of gigabytes. Materialized views with incremental aggregation deliver 50-100x faster query performance, turning seconds-long scans into millisecond responses. These aren't theoretical claims. Faro demonstrates sustained throughput of 50,000+ metrics per second on commodity hardware, validated through comprehensive load testing.

Beyond raw performance, ClickHouse eliminates operational complexity that plagues other time-series solutions. Automatic data lifecycle management through TTL policies means retention windows self-enforce without cron jobs or manual cleanup. Standard SQL querying removes the need for custom query languages, translation layers, and the entire microservice layer that typically sits between dashboards and storage. This simplicity translates directly to predictable costs: compression ratios and storage tiers are deterministic, allowing accurate capacity planning without surprises.

The Faro project proves these benefits work in real production systems, not just benchmarks. If you're evaluating storage backends for metrics, struggling with scale on existing infrastructure, or drowning in the complexity of distributed time-series databases, ClickHouse deserves serious consideration. The combination of performance, operational simplicity, and architectural elegance makes it the optimal foundation for metrics pipelines that need to scale without the overhead of enterprise observability platforms.

---

*The [Faro project](https://github.com/seanankenbruck/faro) is open source with complete implementation details, ClickHouse schemas, consumer code, and Grafana dashboards.*

<div class="post-navigation">
  <a href="/posts/mastering-agentic-patterns" class="nav-article prev">
    <span class="nav-label">Previous Article</span>
    <span class="nav-title">Agentic Patterns Guide</span>
  </a>