# ClipSync - Cross-Platform Clipboard Synchronization

ClipSync is a lightweight, cross-platform clipboard synchronization tool that allows you to share clipboard content between multiple devices on the same network.

## Features

- **Real-time synchronization**: Instantly sync clipboard content between devices
- **Cross-platform**: Works on Windows, macOS, and Linux
- **Local network discovery**: Automatically discovers other ClipSync devices
- **Text and file support**: Sync text content and file paths
- **Secure**: Uses local network communication with no external servers
- **Lightweight**: Minimal resource usage and simple setup

## Installation

### Prerequisites
- Go 1.19 or higher
- Windows 10/11, macOS, or Linux

### Build from source
```bash
git clone https://github.com/DiamondOsas/ClipSync.git
cd ClipSync
go build -o clipsync.exe cmd/crossclip/main.go
```

## Usage

### Basic Usage
```bash
# Run with default settings (uses hostname as device ID, port 8080)
./clipsync.exe

# Run with custom device ID and port
./clipsync.exe -device "MyPC" -port 8080
```

### Command Line Options
- `-device`: Unique device identifier (default: hostname)
- `-port`: Port for communication (default: 8080)

## How It Works

1. **Discovery**: Each ClipSync instance broadcasts its presence on the local network using mDNS
2. **Connection**: Devices automatically connect to each other when discovered
3. **Synchronization**: When clipboard content changes on one device, it's sent to all connected devices
4. **Update**: Receiving devices update their local clipboard with the new content

## Architecture

The application consists of several key components:

- **Clipboard Manager**: Monitors and manages local clipboard changes
- **Discovery Service**: Handles device discovery using mDNS
- **Network Layer**: Manages TCP connections between devices
- **Protocol**: Defines message formats for clipboard synchronization
- **Sync Manager**: Coordinates all components and manages the synchronization process

## Security

- All communication happens on the local network
- No external servers or internet connection required
- Devices are identified by unique device IDs
- No encryption currently implemented (designed for trusted local networks)

## Development

### Project Structure
```
ClipSync/
├── cmd/crossclip/main.go    # Main application entry point
├── internal/
│   ├── clipboard/          # Clipboard management
│   ├── discovery/          # Device discovery
│   ├── network/            # Network communication
│   ├── protocol/           # Message protocol
│   └── sync/               # Synchronization logic
├── go.mod                  # Go module file
└── README.md              # This file
```

### Building
```bash
# Build for current platform
go build -o clipsync.exe cmd/crossclip/main.go

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o clipsync.exe cmd/crossclip/main.go

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o clipsync cmd/crossclip/main.go

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o clipsync cmd/crossclip/main.go
```

## Troubleshooting

### Common Issues

1. **Devices not discovering each other**
   - Ensure all devices are on the same network
   - Check firewall settings to allow mDNS traffic
   - Verify port 8080 is not blocked

2. **Clipboard not syncing**
   - Check that ClipSync is running on all devices
   - Verify devices are connected (check logs)
   - Ensure clipboard content is text-based

3. **Permission issues**
   - On Windows: Run as administrator for clipboard access
   - On macOS: Grant clipboard access in System Preferences
   - On Linux: Install xclip or xsel for clipboard support

### Debug Mode
Run with verbose logging:
```bash
./clipsync.exe -device "MyPC" -port 8080 -debug
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Roadmap

- [ ] File content synchronization
- [ ] End-to-end encryption
- [ ] Configuration file support
- [ ] System tray integration
- [ ] Mobile app support (iOS/Android)
- [ ] Clipboard history
- [ ] Selective sync (choose which devices to sync with)
