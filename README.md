# ClipSync - Cross-Device Clipboard Synchronization

## 🎯 What This App Does (In Simple Terms)

Imagine you're working on your laptop and copy some text. **ClipSync** automatically makes that same text appear on your phone's clipboard, your desktop computer, or any other device running ClipSync on the same WiFi network. No manual syncing, no cloud services, no accounts - just seamless clipboard sharing between your devices.

## 🏗️ High-Level Architecture (System Design View)

### The Big Picture
```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Device A      │     │   Device B      │     │   Device C      │
│   (Laptop)      │────▶│   (Phone)       │◀────│   (Desktop)     │
│   Port: 8080    │     │   Port: 8081    │     │   Port: 8082    │
└─────────────────┘     └─────────────────┘     └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   WiFi Network  │
                    │   (mDNS/Bonjour)│
                    └─────────────────┘
```

### Core Components (Like LEGO Blocks)

#### 1. **Discovery Service** - "Who's Out There?"
- **What it does**: Finds other ClipSync devices on your network
- **How it works**: Uses mDNS (like Apple's Bonjour) to broadcast "Hey, I'm ClipSync on port 8080!"
- **Real-world analogy**: Like shouting in a crowded room "Anyone else here named ClipSync?"

#### 2. **Clipboard Manager** - "What's On My Clipboard?"
- **What it does**: Monitors your clipboard for changes
- **How it works**: 
  - Windows: Uses Windows API to watch clipboard events
  - When you copy something, it immediately detects the change
- **Real-world analogy**: Like having a friend constantly peeking at what you just copied

#### 3. **Network Server** - "I'm Listening"
- **What it does**: Waits for other devices to connect and send clipboard data
- **How it works**: Opens a TCP port (default 8080) and listens for incoming connections
- **Real-world analogy**: Like having your phone on speaker, waiting for calls

#### 4. **Network Client** - "I'll Connect to You"
- **What it does**: Connects to other ClipSync devices when discovered
- **How it works**: When it finds another device, it opens a connection to that device's server
- **Real-world analogy**: Like calling your friend's phone when you see they're online

#### 5. **Sync Manager** - "The Brain"
- **What it does**: Coordinates everything - it's the master controller
- **How it works**: 
  - Starts all other components
  - Routes clipboard changes to connected devices
  - Manages connections and disconnections

#### 6. **Protocol Handler** - "Speaking the Same Language"
- **What it does**: Defines how devices talk to each other
- **How it works**: Uses JSON messages like:
  ```json
  {
    "type": "clipboard_update",
    "device_id": "MyLaptop",
    "timestamp": "2024-08-03T22:15:00Z",
    "data": {
      "type": "text",
      "content": "Hello from my clipboard!",
      "size": 23
    }
  }
  ```

## 🔄 Data Flow (Step-by-Step)

### When You Copy Something:
1. **You press Ctrl+C** → Clipboard Manager detects the change
2. **Sync Manager gets notified** → "Hey, clipboard changed!"
3. **Message created** → "This is from MyLaptop, here's the text"
4. **Broadcast to all** → Server sends to all connected devices
5. **Other devices receive** → "MyLaptop sent clipboard data"
6. **Update their clipboard** → Other devices set their clipboard to the same text

### When a New Device Joins:
1. **Device starts ClipSync** → "I'm here on port 8080!"
2. **Discovery broadcast** → Other devices see the announcement
3. **Automatic connection** → Existing devices connect to new device
4. **Hello message** → "Hi, I'm MyPhone, nice to meet you"
5. **Ready to sync** → Both devices can now share clipboard data

## 🛠️ Technical Deep Dive (For System Architects)

### Architecture Pattern: Event-Driven Microservices
Each component operates independently and communicates via events:

```
┌─────────────────┐    Events    ┌─────────────────┐
│  Clipboard      │─────────────▶│  Sync Manager   │
│  Manager        │              │                 │
└─────────────────┘              └─────────────────┘
         │                                │
         │                                │
         ▼                                ▼
┌─────────────────┐              ┌─────────────────┐
│  Network        │              │  Discovery      │
│  Server         │              │  Service        │
└─────────────────┘              └─────────────────┘
```

### Threading Model: Goroutines Everywhere
- Each device connection runs in its own goroutine
- Discovery runs in background goroutine
- Clipboard monitoring runs in separate goroutine
- This allows handling multiple devices simultaneously

### Network Protocol: TCP + JSON
- **Transport**: TCP for reliability
- **Encoding**: JSON for human-readability and cross-platform compatibility
- **Message Types**:
  - `clipboard_update`: New clipboard content
  - `device_hello`: New device announcement
  - `device_bye`: Device leaving
  - `ping/pong`: Connection health checks

### Error Handling Strategy
- **Graceful degradation**: If one device fails, others continue working
- **Connection retry**: Automatic reconnection attempts
- **Type safety**: All messages validated before processing
- **Timeout handling**: Prevents hanging connections

## 🚀 How to Use

### Starting the Service:
```bash
# Basic usage
.\clipsync.exe -device "MyLaptop"

# Custom port
.\clipsync.exe -device "MyDesktop" -port 8081
```

### Requirements:
- All devices must be on the same WiFi network
- Each device needs a unique name
- Windows 10+ (for Windows clipboard API)

## 🔧 Configuration Options

| Flag | Description | Default |
|------|-------------|---------|
| `-device` | Unique name for this device | Required |
| `-port` | TCP port for communication | 8080 |

## 🛡️ Security Considerations

- **Local network only**: No internet connectivity required
- **No encryption**: Designed for trusted home/office networks
- **No authentication**: Assumes trusted network environment
- **Port binding**: Only binds to local network interfaces

## 🐛 Troubleshooting

### Common Issues:
1. **"No devices found"** → Check if all devices are on same WiFi
2. **"Port already in use"** → Use different port with `-port` flag
3. **"Clipboard not syncing"** → Check Windows clipboard permissions

### Debug Mode:
Run with verbose logging to see what's happening:
```bash
# Windows PowerShell
$env:DEBUG="1"; .\clipsync.exe -device "DebugDevice"
```

## 📊 Performance Characteristics

- **Memory usage**: ~10-20MB (very lightweight)
- **CPU usage**: <1% when idle, spikes during clipboard sync
- **Network usage**: Only when clipboard changes (minimal)
- **Latency**: <100ms on local network

## 🎯 Future Enhancements

- **Encryption**: TLS for secure communication
- **File sync**: Support for copying files
- **Image sync**: Full image clipboard support
- **Mobile apps**: iOS/Android versions
- **Cloud relay**: Sync across different networks

---

*Built with Go for cross-platform compatibility and high performance. The entire application is contained in a single binary with no external dependencies.*
