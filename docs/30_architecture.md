# Architecture

The Integration Connector Agent follows a modular, pipeline-based architecture that enables flexible data flow from various
sources through configurable processors to multiple destinations. This document provides a comprehensive overview of the
system architecture and data flow patterns.

## 🏗️ High-Level Architecture

The Integration Connector Agent consists of three main components that work together to form data integration pipelines:

### Core Components

- **Sources**: External systems that generate or provide data
- **Processors**: Data transformation and filtering components  
- **Sinks**: Destination systems where processed data is stored or forwarded

![High-Level Architecture](./img/architecture.excalidraw.svg)

### Agent Components

- **Configuration Manager**: Loads and validates pipeline configurations
- **Source Manager**: Manages connections to external data sources
- **Pipeline Orchestrator**: Coordinates data flow through processor chains
- **Sink Manager**: Handles data delivery to destination systems
- **Event Bus**: Internal messaging system for component communication
- **Health Monitor**: Monitors component health and pipeline status

## 🔄 Data Flow Pipeline

The following diagram illustrates the complete data flow from source to sink:

```mermaid
graph TD
    A[External Source] -->|Raw Events| B[Source Adapter]
    B -->|Structured Events| C[Event Bus]
    C -->|Event Stream| D[Pipeline Orchestrator]
    
    D -->|Event| E[Processor 1: Filter]
    E -->|Filtered Event| F[Processor 2: Mapper]
    F -->|Transformed Event| G[Processor N: Custom]
    
    G -->|Processed Event| H[Sink Adapter]
    H -->|Formatted Data| I[Destination System]
    
    D -.->|Parallel Processing| J[Pipeline 2]
    D -.->|Parallel Processing| K[Pipeline N]
    
    style A fill:#e1f5fe
    style I fill:#f3e5f5
    style C fill:#fff3e0
    style D fill:#e8f5e8
```

### Detailed Flow Steps

- **Data Ingestion**
  - Source adapters connect to external systems
  - Raw data is normalized into internal event format
  - Events are published to the internal event bus

- **Pipeline Processing**
  - Pipeline Orchestrator receives events from the event bus
  - Each pipeline processes events through its configured processor chain
  - Processors can filter, transform, enrich, or validate data

- **Data Output**
  - Processed events are sent to configured sinks
  - Sink adapters format data according to destination requirements
  - Data is delivered to target systems with retry and error handling

## 📊 Component Architecture

### Source Components

```mermaid
graph TB
    subgraph "Source Layer"
        A[GitHub Source]
        B[GitLab Source]
        C[Jira Source]
        D[Azure Source]
        E[Confluence Source]
        F[AWS Source]
        G[GCP Source]
        H[JBoss Source]
    end
    
    subgraph "Source Framework"
        I[HTTP Client]
        J[Webhook Server]
        K[Polling Manager]
        L[Authentication Manager]
    end
    
    A --> I
    B --> I
    C --> I
    D --> I
    E --> I
    F --> I
    G --> I
    H --> I
    
    A --> J
    B --> J
    
    A --> K
    B --> K
    C --> K
    D --> K
    E --> K
    F --> K
    G --> K
    H --> K
    
    A --> L
    B --> L
    C --> L
    D --> L
    E --> L
    F --> L
    G --> L
    H --> L
```

### Processor Components

```mermaid
graph TB
    subgraph "Processor Layer"
        A[Filter Processor]
        B[Mapper Processor]
        C[RPC Plugin Processor]
        D[Cloud Vendor Aggregator]
    end
    
    subgraph "Processing Framework"
        E[CEL Expression Engine]
        F[Template Engine]
        G[RPC Client]
        H[Aggregation Engine]
    end
    
    A --> E
    B --> F
    C --> G
    D --> H
    
    style A fill:#ffeb3b
    style B fill:#4caf50
    style C fill:#2196f3
    style D fill:#ff9800
```

### Sink Components

```mermaid
graph TB
    subgraph "Sink Layer"
        A[Console Catalog Sink]
        B[MongoDB Sink]
        C[CRUD Service Sink]
        D[Kafka Sink]
    end
    
    subgraph "Sink Framework"
        E[HTTP Client]
        F[Database Driver]
        G[Message Queue Client]
        H[Retry Manager]
    end
    
    A --> E
    B --> F
    C --> E
    D --> G
    
    A --> H
    B --> H
    C --> H
    D --> H
```

## 🔧 Configuration Architecture

The agent uses a hierarchical configuration structure that defines integrations, pipelines, and components:

```yaml
Configuration:
  ├── Integrations[]
  │   ├── Source
  │   │   ├── Type (github, gitlab, jira, etc.)
  │   │   ├── Connection Settings
  │   │   └── Authentication
  │   └── Pipelines[]
  │       ├── Processors[]
  │       │   ├── Type (filter, mapper, rpc, etc.)
  │       │   └── Configuration
  │       └── Sinks[]
  │           ├── Type (console-catalog, mongo, etc.)
  │           └── Connection Settings
```

### Configuration Flow

```mermaid
graph TD
    A[config.json] -->|Load| B[Configuration Parser]
    B -->|Validate| C[Schema Validator]
    C -->|Create| D[Source Instances]
    C -->|Create| E[Pipeline Instances]
    C -->|Create| F[Sink Instances]
    
    D -->|Register| G[Source Manager]
    E -->|Register| H[Pipeline Orchestrator]
    F -->|Register| I[Sink Manager]
    
    G -->|Events| J[Event Bus]
    H -->|Subscribe| J
    I -->|Subscribe| H
```

## 🚀 Deployment Architecture

### Standalone Deployment

```mermaid
graph TB
    subgraph "Container Environment"
        A[Integration Connector Agent]
        B[Configuration Volume]
        C[Logs Volume]
    end
    
    subgraph "External Sources"
        D[GitHub API]
        E[GitLab API]
        F[Jira API]
        G[Azure APIs]
    end
    
    subgraph "Destination Systems"
        H[Mia-Platform Console]
        I[MongoDB]
        J[Kafka]
    end
    
    A --> D
    A --> E
    A --> F
    A --> G
    
    A --> H
    A --> I
    A --> J
    
    B --> A
    A --> C
```

### Kubernetes Deployment

```mermaid
graph TB
    subgraph "Kubernetes Cluster"
        subgraph "Namespace: integration-connector"
            A[Deployment: connector-agent]
            B[ConfigMap: agent-config]
            C[Secret: credentials]
            D[Service: agent-service]
            E[PVC: logs-storage]
        end
        
        subgraph "Monitoring"
            F[ServiceMonitor]
            G[PrometheusRule]
        end
    end
    
    A --> B
    A --> C
    A --> E
    D --> A
    F --> A
    G --> A
    
    subgraph "External Systems"
        H[External APIs]
        I[Target Sinks]
    end
    
    A -.->|HTTPS| H
    A -.->|HTTPS| I
```

## 🔄 Event Processing Model

### Event Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Received: Source Event
    Received --> Queued: Add to Event Bus
    Queued --> Processing: Pipeline Pickup
    Processing --> Filtered: Apply Filters
    Filtered --> Transformed: Apply Mappers
    Transformed --> Enriched: Apply Processors
    Enriched --> Delivered: Send to Sinks
    Delivered --> [*]: Complete
    
    Processing --> Dropped: Filter Excludes
    Dropped --> [*]: Discard
    
    Processing --> Failed: Error Occurred
    Failed --> Retry: Retryable Error
    Failed --> DeadLetter: Max Retries
    Retry --> Processing: Retry Attempt
    DeadLetter --> [*]: Manual Intervention
```

### Parallel Processing

The agent supports parallel processing across multiple dimensions:

- **Source Parallelism**: Multiple source instances can run concurrently
- **Pipeline Parallelism**: Each source can have multiple independent pipelines
- **Processor Parallelism**: Processors within a pipeline can be chained or parallel
- **Sink Parallelism**: Each pipeline can send data to multiple sinks simultaneously

```mermaid
graph TD
    A[Source Event] --> B[Event Bus]
    
    B --> C[Pipeline 1]
    B --> D[Pipeline 2]
    B --> E[Pipeline N]
    
    C --> F[Processor Chain 1]
    D --> G[Processor Chain 2]
    E --> H[Processor Chain N]
    
    F --> I[Sink A]
    F --> J[Sink B]
    G --> K[Sink C]
    H --> L[Sink D]
    H --> M[Sink E]
```

## 📈 Scalability and Performance

### Horizontal Scaling

- **Multi-Instance Deployment**: Deploy multiple agent instances for increased throughput
- **Source Partitioning**: Distribute sources across different agent instances
- **Load Balancing**: Use external load balancers for webhook endpoints

### Vertical Scaling

- **Memory Optimization**: Configurable buffer sizes and batch processing
- **CPU Optimization**: Parallel processor execution and async I/O
- **Network Optimization**: Connection pooling and keep-alive connections

### Performance Monitoring

The agent exposes metrics for monitoring and optimization:

- **Source Metrics**: Event ingestion rates, connection health
- **Pipeline Metrics**: Processing latency, throughput, error rates
- **Sink Metrics**: Delivery success rates, retry attempts, latency

## 🔒 Security Architecture

### Authentication & Authorization

```mermaid
graph TD
    A[Agent] -->|API Keys| B[Source APIs]
    A -->|Service Account| C[Mia-Platform Console]
    A -->|Connection Strings| D[Databases]
    A -->|Certificates| E[Message Queues]
    
    F[Secret Manager] --> A
    G[Environment Variables] --> A
    H[Config Files] --> A
```

### Security Layers

- **Transport Security**: TLS encryption for all external communications
- **Authentication**: Multiple authentication methods (API keys, OAuth, service accounts)
- **Secret Management**: Secure storage and rotation of credentials
- **Network Security**: VPC/firewall rules for network isolation
- **Runtime Security**: Container security and resource limits

## 🛠️ Extensibility

The agent is designed for extensibility through well-defined interfaces:

### Plugin Architecture

- **Source Plugins**: Implement new data sources
- **Processor Plugins**: Add custom data transformation logic
- **Sink Plugins**: Support new destination systems

### Custom Components

- **RPC Processors**: External processing services via gRPC
- **Custom Authentication**: Pluggable authentication providers
- **Custom Formats**: Support for additional data formats and protocols

This architecture ensures the Integration Connector Agent can adapt to new requirements and integrate with evolving
technology stacks while maintaining reliability and performance.
