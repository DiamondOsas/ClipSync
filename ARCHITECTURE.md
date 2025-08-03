# ClipSync Architecture Diagrams & Visual Guide

## 🎯 System Overview Diagram

```mermaid
graph TB
    subgraph "Device A - MyLaptop"
        A[Sync Manager<br/>:8080]
        A1[Clipboard<br/>Manager]
        A2[Discovery<br/>Service]
        A3[Network<br/>Server]
        A4[Network<br/>Client]
        
        A --> A1
        A --> A2
        A --> A3
        A --> A4
    end
    
    subgraph "Device B - MyPhone"
        B[Sync Manager<br/>:8081]
        B1[Clipboard<br/>Manager]
        B2[Discovery<br/>Service]
        B3[Network<br/>Server]
        B4[Network<br/>Client]
        
        B --> B1
        B --> B2
        B --> B3
        B --> B4
    end
    
    subgraph "WiFi Network"
        MDNS[mDNS/Bonjour<br/>Multicast DNS]
    end
    
    A2 -.->|Broadcast| MDNS
    B2 -.->|Broadcast| MDNS
    
    A4 -->|TCP Connection| B3
    B4 -->|TCP Connection| A3
    
    style A fill:#f9f,stroke:#333
    style B fill:#9f9,stroke:#333
    style MDNS fill:#ff9,stroke:#333
```

## 🔍 Component Interaction Flow

### 1. Discovery Phase
```mermaid
sequenceDiagram
    participant Laptop as Device A (Laptop)
    participant Phone as Device B (Phone)
    participant Network as WiFi Network
    
    Note over Laptop,Phone: Both devices start ClipSync
    
    Laptop->>Network: mDNS broadcast: "ClipSync on 8080"
    Phone->>Network: mDNS broadcast: "ClipSync on 8081"
    
    Phone->>Laptop: TCP connect to :8080
    Laptop->>Phone: TCP connect to :8081
    
    Note over Laptop,Phone: Devices are now connected
```

### 2. Clipboard Sync Flow
```mermaid
sequenceDiagram
    participant User as User
    participant Clipboard as Clipboard Manager
    participant Sync as Sync Manager
    participant Server as Network Server
    participant Client as Network Client
    participant Remote as Remote Device
    
    User->>Clipboard: Ctrl+C "Hello World"
    Clipboard->>Sync: Event: clipboard changed
    Sync->>Sync: Create JSON message
    Sync->>Server: Broadcast to all clients
    
    Server->>Client: Send message via TCP
    Client->>Remote: Receive and process
    Remote->>Remote: Update local clipboard
    
    Note over Remote: Remote device now has<br/>"Hello World" in clipboard
```

## 🏗️ Detailed Component Architecture

### Clipboard Manager Architecture
```mermaid
graph LR
    subgraph "Clipboard Manager"
        CM[Clipboard Manager]
        WC[Windows API<br/>Clipboard Watcher]
        CB[Current<br/>Clipboard]
        CBH[Change<br/>Handlers]
        
        WC -->|detects change| CM
        CM -->|reads| CB
        CM -->|notifies| CBH
    end
    
    style CM fill:#bbf,stroke:#333
    style WC fill:#bfb,stroke:#333
```

### Discovery Service Architecture
```mermaid
graph TD
    subgraph "Discovery Service"
        DS[Discovery Service]
        ZC[Zeroconf/mDNS]
        REG[Service Registry]
        BROW[Browser]
        
        DS -->|register| ZC
        DS -->|announce| REG
        DS -->|discover| BROW
        BROW -->|find devices| ZC
    end
    
    style DS fill:#fbb,stroke:#333
    style ZC fill:#bfb,stroke:#333
```

### Network Layer Architecture
```mermaid
graph LR
    subgraph "Network Layer"
        NS[Network Server]
        NC[Network Client]
        CONN[TCP Connections]
        MSG[Message Queue]
        
        NS -->|accepts| CONN
        CONN -->|sends| MSG
        NC -->|connects| CONN
        CONN -->|receives| MSG
    end
    
    style NS fill:#fbf,stroke:#333
    style NC fill:#bff,stroke:#333
```

## 🔄 State Machine Diagrams

### Device Connection States
```mermaid
stateDiagram-v2
    [*] --> Starting: Device starts
    Starting --> Discovering: Begin mDNS broadcast
    Discovering --> Connecting: Found peer device
    Connecting --> Connected: TCP handshake successful
    Connected --> Syncing: Exchange hello messages
    Syncing --> Connected: Ready for clipboard sync
    
    Connected --> Disconnected: Connection lost
    Disconnected --> Connecting: Retry connection
    Connecting --> Failed: Max retries exceeded
    Failed --> [*]: Shutdown
    
    Syncing --> Disconnected: Peer disconnects
```

### Message Processing States
```mermaid
stateDiagram-v2
    [*] --> Idle: Waiting for messages
    Idle --> Receiving: TCP data received
    Receiving --> Parsing: Decode JSON
    Parsing --> Validating: Check message format
    
    Validating --> Processing: Valid message
    Validating --> Error: Invalid message
    Error --> Idle: Log and continue
    
    Processing --> ClipboardUpdate: Type = clipboard_update
    Processing --> DeviceHello: Type = device_hello
    Processing --> DeviceBye: Type = device_bye
    
    ClipboardUpdate --> UpdateClipboard: Set local clipboard
    DeviceHello --> LogHello: Log new device
    DeviceBye --> RemoveDevice: Clean up connection
    
    UpdateClipboard --> Idle
    LogHello --> Idle
    RemoveDevice --> Idle
```

## 📊 Data Flow Architecture

### Message Structure
```json
{
  "header": {
    "type": "clipboard_update|device_hello|device_bye|ping|pong",
    "device_id": "unique_device_name",
    "timestamp": "2024-08-03T22:15:00.000Z"
  },
  "payload": {
    "clipboard_data": {
      "type": "text|image|files",
      "content": "actual clipboard content",
      "size": 1234,
      "metadata": {
        "format": "utf-8",
        "encoding": "base64"
      }
    }
  }
}
```

### Threading Model
```mermaid
graph TD
    subgraph "Goroutines"
        M[Main Goroutine]
        CM[Clipboard Monitor]
        DS[Discovery Service]
        NS[Network Server]
        NC[Network Client]
        MH[Message Handlers]
        
        M -->|starts| CM
        M -->|starts| DS
        M -->|starts| NS
        
        DS -->|spawns| NC
        NS -->|spawns| MH
        
        CM -.->|events| M
        MH -.->|messages| M
    end
    
    style M fill:#ff9,stroke:#333
    style CM fill:#9ff,stroke:#333
    style DS fill:#f9f,stroke:#333
```

## 🛡️ Error Handling Architecture

### Resilience Patterns
```mermaid
graph TD
    subgraph "Error Handling"
        ERR[Error Detection]
        RETRY[Retry Logic]
        FALLBACK[Fallback Strategy]
        LOG[Logging System]
        
        ERR -->|connection lost| RETRY
        ERR -->|invalid message| LOG
        ERR -->|clipboard error| FALLBACK
        
        RETRY -->|max retries| FALLBACK
        FALLBACK -->|skip device| LOG
    end
    
    style ERR fill:#fbb,stroke:#333
    style RETRY fill:#bfb,stroke:#333
```

## 🔧 Configuration Architecture

### Settings Hierarchy
```mermaid
graph TD
    subgraph "Configuration"
        CLI[Command Line Args]
        ENV[Environment Vars]
        DEFAULT[Default Values]
        
        CLI -->|overrides| ENV
        ENV -->|overrides| DEFAULT
        
        CLI -->|device name| CONFIG[Final Config]
        ENV -->|port number| CONFIG
        DEFAULT -->|fallback| CONFIG
    end
    
    style CLI fill:#bbf,stroke:#333
    style ENV fill:#fbf,stroke:#333
```

## 📈 Performance Architecture

### Resource Usage Patterns
```mermaid
graph LR
    subgraph "Performance Metrics"
        CPU[CPU Usage<br/><1% idle]
        MEM[Memory<br/>10-20MB]
        NET[Network<br/>Minimal]
        LAT[Latency<br/><100ms]
        
        CPU -->|spikes during sync| NET
        MEM -->|constant| LAT
    end
    
    style CPU fill:#fbb,stroke:#333
    style MEM fill:#bfb,stroke:#333
```

## 🎯 Deployment Architecture

### Single Binary Distribution
```mermaid
graph TD
    subgraph "Deployment"
        BUILD[Go Build]
        BIN[Single Binary<br/>clipsync.exe]
        DIST[Cross-Platform<br/>Distribution]
        
        BUILD -->|compiles to| BIN
        BIN -->|no dependencies| DIST
        DIST -->|Windows| WIN[Windows 10+]
        DIST -->|Linux| LIN[Linux x64]
        DIST -->|macOS| MAC[macOS x64]
    end
    
    style BUILD fill:#9ff,stroke:#333
    style BIN fill:#ff9,stroke:#333
```

## 🔍 Monitoring Architecture

### Health Check System
```mermaid
graph LR
    subgraph "Monitoring"
        HC[Health Check]
        METRICS[Metrics Collection]
        ALERTS[Alert System]
        
        HC -->|ping devices| METRICS
        METRICS -->|threshold| ALERTS
        ALERTS -->|log| LOGS[System Logs]
    end
    
    style HC fill:#fbf,stroke:#333
    style METRICS fill:#bbf,stroke:#333
```

---

*These diagrams represent the complete system architecture of ClipSync, designed for maximum clarity and understanding at all technical levels.*
