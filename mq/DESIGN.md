# Message Queue - Design Document

A distributed, durable message queue with Kafka-like semantics built in Go.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           CLIENT LAYER                                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                      │
│  │ Go Producer  │  │ Go Consumer  │  │ CLI Tool     │                      │
│  │ (gRPC stream)│  │ (gRPC stream)│  │ (admin ops)  │                      │
│  └──────────────┘  └──────────────┘  └──────────────┘                      │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                              gRPC / Protobuf
                                    │
┌─────────────────────────────────────────────────────────────────────────────┐
│                           BROKER CLUSTER                                     │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    RAFT CONSENSUS LAYER                              │   │
│  │  • Leader election for each partition                                │   │
│  │  • Log replication                                                   │   │
│  │  • Membership changes                                                │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐            │
│  │    Broker 1     │  │    Broker 2     │  │    Broker 3     │            │
│  │  (Port 9001)    │  │  (Port 9002)    │  │  (Port 9003)    │            │
│  │                 │  │                 │  │                 │            │
│  │ ┌─────────────┐ │  │ ┌─────────────┐ │  │ ┌─────────────┐ │            │
│  │ │ Partition 0 │ │  │ │ Partition 1 │ │  │ │ Partition 2 │ │            │
│  │ │   (Leader)  │ │  │ │   (Leader)  │ │  │ │   (Leader)  │ │            │
│  │ └─────────────┘ │  │ └─────────────┘ │  │ └─────────────┘ │            │
│  │ ┌─────────────┐ │  │ ┌─────────────┐ │  │ ┌─────────────┐ │            │
│  │ │ Partition 1 │ │  │ │ Partition 2 │ │  │ │ Partition 0 │ │            │
│  │ │ (Follower)  │ │  │ │ (Follower)  │ │  │ │ (Follower)  │ │            │
│  │ └─────────────┘ │  │ └─────────────┘ │  │ └─────────────┘ │            │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘            │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
┌─────────────────────────────────────────────────────────────────────────────┐
│                           STORAGE LAYER                                      │
│                                                                              │
│   data/                                                                      │
│   ├── topics/                                                               │
│   │   └── orders/                                                           │
│   │       ├── partition-0/                                                  │
│   │       │   ├── 00000000000000000000.log      ← Message segments          │
│   │       │   ├── 00000000000000000000.index    ← Sparse offset index       │
│   │       │   ├── 00000000000000000000.timeindex ← Time-based index         │
│   │       │   └── 00000000000000100000.log      ← Rolled segment            │
│   │       └── partition-1/                                                  │
│   ├── __consumer_offsets/       ← Internal offset commit topic              │
│   └── cluster-metadata/         ← Raft logs & snapshots                     │
│       ├── raft.log                                                         │
│       └── raft.snap                                                         │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. Storage Engine (`/internal/storage`)

| File | Purpose |
|------|---------|
| `wal.go` | Write-ahead log with segment management |
| `segment.go` | Individual segment file (log + index) |
| `index.go` | Sparse offset index (mmap'd) + time index |
| `reader.go` | Sequential and random message reads |
| `compactor.go` | Log compaction / retention policies |

**Key Features:**
- Sequential write path (O(1) append)
- Memory-mapped indexes for O(log n) random reads
- Configurable segment size (default 1GB)
- Zero-copy reads where possible
- CRC32 checksums per message

### 2. Broker Core (`/internal/broker`)

| File | Purpose |
|------|---------|
| `broker.go` | Main broker struct, lifecycle |
| `topic.go` | Topic and partition management |
| `partition.go` | Partition leader/follower logic |
| `replicator.go` | ISRs, replication factor, acks |
| `offset_tracker.go` | High watermark, log end offset |

**Replication Model:**
- Sync replication to ISR (In-Sync Replicas)
- Configurable `acks` (0, 1, all)
- Min ISR for writes
- Follower fetching (pull-based replication)

### 3. Raft Consensus (`/internal/raft`)

| File | Purpose |
|------|---------|
| `node.go` | Raft node state machine |
| `log.go` | Raft log entries |
| `state.go` | Leader/Follower/Candidate states |
| `election.go` | Leader election timeout, voting |
| `snapshot.go` | Log compaction snapshots |

**Recommendation:** Build simplified Raft for educational purposes.

### 4. gRPC Service (`/proto`, `/internal/api`)

**Protobuf Definitions:**

```protobuf
// proto/messagequeue.proto

service MessageQueue {
  // Producer APIs
  rpc Produce(ProduceRequest) returns (ProduceResponse);
  rpc ProduceStream(stream ProduceRequest) returns (ProduceResponse);
  
  // Consumer APIs  
  rpc Subscribe(SubscribeRequest) returns (stream Message);
  rpc CommitOffset(CommitOffsetRequest) returns (CommitOffsetResponse);
  rpc Seek(SeekRequest) returns (SeekResponse);
  
  // Admin APIs
  rpc CreateTopic(CreateTopicRequest) returns (CreateTopicResponse);
  rpc DeleteTopic(DeleteTopicRequest) returns (DeleteTopicResponse);
  rpc DescribeCluster(DescribeClusterRequest) returns (DescribeClusterResponse);
  
  // Cluster Internal
  rpc Replicate(stream ReplicateRequest) returns (stream ReplicateResponse);
  rpc Heartbeat(HeartbeatRequest) returns (HeartbeatResponse);
  rpc Vote(VoteRequest) returns (VoteResponse);
}

message Message {
  bytes key = 1;
  bytes value = 2;
  map<string, string> headers = 3;
  int64 timestamp = 4;
  int64 offset = 5;
  int32 partition = 6;
}

message ProduceRequest {
  string topic = 1;
  bytes key = 2;
  bytes value = 3;
  map<string, string> headers = 4;
  int32 acks = 5;  // 0, 1, -1 (all)
}

message SubscribeRequest {
  string group_id = 1;
  string topic = 2;
  int64 start_offset = 3;  // -1 = latest, -2 = earliest
}
```

### 5. Consumer Groups (`/internal/consumer`)

| File | Purpose |
|------|---------|
| `group.go` | Consumer group state machine |
| `coordinator.go` | Group coordinator per broker |
| `rebalancer.go` | Partition assignment strategies |
| `offset_commit.go` | Offset commit log to internal topic |

**Rebalancing Protocol:**
1. JoinGroup → Coordinator assigns member IDs
2. SyncGroup → Assign partitions to members
3. Heartbeat → Keep session alive
4. LeaveGroup → Clean leave

### 6. Client Libraries (`/client`)

| File | Purpose |
|------|---------|
| `producer.go` | Batched, async producer with retries |
| `consumer.go` | Consumer group member with auto-commit |
| `admin.go` | Topic CRUD, cluster metadata |

---

## Project Structure

```
mq/
├── cmd/
│   ├── broker/           # Broker binary
│   │   └── main.go
│   ├── cli/              # Admin CLI tool
│   │   └── main.go
│   └── benchmark/        # Performance testing
│       └── main.go
├── proto/
│   └── messagequeue.proto
├── internal/
│   ├── storage/          # WAL, segments, indexes
│   │   ├── wal.go
│   │   ├── segment.go
│   │   ├── index.go
│   │   ├── reader.go
│   │   └── compactor.go
│   ├── broker/           # Core broker logic
│   │   ├── broker.go
│   │   ├── topic.go
│   │   ├── partition.go
│   │   ├── replicator.go
│   │   └── offset_tracker.go
│   ├── raft/             # Simplified Raft
│   │   ├── node.go
│   │   ├── state.go
│   │   ├── election.go
│   │   └── snapshot.go
│   ├── consumer/         # Consumer group coordination
│   │   ├── group.go
│   │   ├── coordinator.go
│   │   ├── rebalancer.go
│   │   └── offset_commit.go
│   └── api/              # gRPC service implementations
│       ├── produce.go
│       ├── consume.go
│       ├── admin.go
│       └── internal.go
├── client/               # Go client library
│   ├── producer.go
│   ├── consumer.go
│   ├── admin.go
│   └── client.go
├── pkg/
│   ├── protocol/         # Generated protobuf code
│   └── codec/            # Message encoding/decoding
├── configs/
│   └── broker.yaml       # Sample configuration
├── scripts/
│   ├── start-cluster.sh  # Start 3-node cluster
│   └── benchmark.sh      # Run performance tests
├── go.mod
├── Makefile
└── README.md
```

---

## Implementation Phases

| Phase | Scope | Complexity | Est. Time |
|-------|-------|------------|-----------|
| **1. Storage** | WAL, segments, indexes, reads | High | 2-3 days |
| **2. Single Broker** | gRPC API, produce/consume on single node | Medium | 1-2 days |
| **3. Consumer Groups** | Group coordination, rebalancing, offset commits | Medium-High | 2 days |
| **4. Raft Consensus** | Leader election, log replication | Very High | 3-4 days |
| **5. Replication** | Partition replication, ISRs, failover | High | 2-3 days |
| **6. Client Libs** | Producer, consumer, admin client | Medium | 1-2 days |
| **7. CLI & Benchmark** | Admin tool, throughput tests | Low | 1 day |

---

## Message Format (Binary)

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Offset: 8 bytes                                                              │
├──────────────────────────────────────────────────────────────────────────────┤
│ CRC32: 4 bytes                                                               │
├──────────────────────────────────────────────────────────────────────────────┤
│ Length: 4 bytes (total message size excluding offset)                        │
├──────────────────────────────────────────────────────────────────────────────┤
│ Attributes: 1 byte (compression, timestamp type, etc.)                       │
├──────────────────────────────────────────────────────────────────────────────┤
│ Timestamp: 8 bytes                                                           │
├──────────────────────────────────────────────────────────────────────────────┤
│ Key Length: 4 bytes │ Key: N bytes                                          │
├──────────────────────────────────────────────────────────────────────────────┤
│ Value Length: 4 bytes │ Value: N bytes                                      │
├──────────────────────────────────────────────────────────────────────────────┤
│ Headers Count: 4 bytes                                                       │
│ ┌────────────────────────────────────────────────────────────────────────┐  │
│ │ Header Key Length: 2 bytes │ Header Key: N bytes                       │  │
│ │ Header Value Length: 2 bytes │ Header Value: N bytes                   │  │
│ │ ... (repeat for each header)                                           │  │
│ └────────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Index Format

### Offset Index (Sparse - every 4KB of log)

```
┌────────────────────────────────────────────┐
│ Base Offset: 8 bytes                       │
├────────────────────────────────────────────┤
│ Relative Offset: 4 bytes                   │
│ (actual offset = base_offset + relative)   │
├────────────────────────────────────────────┤
│ File Position: 4 bytes                     │
└────────────────────────────────────────────┘

Total: 8 bytes per entry, sparse (every ~4096 bytes of log data)
Enables binary search for offset → file position
```

### Time Index

```
┌────────────────────────────────────────────┐
│ Timestamp: 8 bytes                         │
├────────────────────────────────────────────┤
│ Relative Offset: 4 bytes                   │
└────────────────────────────────────────────┘

Enables time-based message seeking
```

---

## Configuration

```yaml
# configs/broker.yaml

broker:
  id: 1
  host: "0.0.0.0"
  port: 9001
  data_dir: "./data"

storage:
  segment_size: 1073741824  # 1GB
  index_interval: 4096      # bytes between index entries
  retention_ms: 604800000   # 7 days
  retention_bytes: -1       # unlimited
  compaction_enabled: false

cluster:
  seed_nodes:
    - "localhost:9001"
    - "localhost:9002"
    - "localhost:9003"
  election_timeout_ms: 1000
  heartbeat_interval_ms: 200

replication:
  default_replication_factor: 3
  min_insync_replicas: 2
  replica_fetch_max_bytes: 1048576

consumer:
  session_timeout_ms: 10000
  heartbeat_interval_ms: 3000
  max_poll_interval_ms: 300000
  auto_commit_interval_ms: 5000

log:
  level: "info"
  format: "json"
```

---

## Key Technical Decisions

### 1. Backpressure & Flow Control

- Consumer: `MAX_IN_FLIGHT_REQUESTS` per stream
- Producer: Batch accumulation with timeout
- Broker: Connection-level rate limiting

### 2. Idempotent Producer

- Sequence numbers per partition per producer ID
- Deduplication window (sliding, configurable)
- Producer ID via `InitProducerId` RPC

### 3. Transactions (Future)

- Transaction coordinator
- Transaction log internal topic
- Two-phase commit for consumer groups

### 4. Performance Targets

| Metric | Target |
|--------|--------|
| Produce throughput (single partition) | > 100k msg/sec |
| Produce latency (p99, acks=1) | < 10ms |
| Consume throughput | > 200k msg/sec |
| End-to-end latency | < 20ms |

---

## Getting Started (After Implementation)

```bash
# Build
make build

# Start 3-node cluster
./scripts/start-cluster.sh

# Create topic
./bin/cli topics create orders --partitions 3 --replication-factor 3

# Produce messages
./bin/cli produce orders --key "order-1" --value '{"item": "widget", "qty": 100}'

# Consume messages
./bin/cli consume orders --group "order-processor"

# Benchmark
./bin/benchmark --topic orders --messages 100000
```

---

## References

- [Kafka Internals](https://kafka.apache.org/documentation/#design)
- [Raft Paper](https://raft.github.io/raft.pdf)
- [Designing Data-Intensive Applications](https://dataintensive.net/)
- [Pulsar Architecture](https://pulsar.apache.org/docs/en/concepts-architecture-overview/)
