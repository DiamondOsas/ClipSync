# How ClipSync Works

ClipSync is a LAN-based clipboard synchronization tool that allows sharing clipboard content between devices on the same network. Here's how it works:

## Core Components

1. **Service Discovery**: Uses mDNS (multicast DNS) for automatic device discovery
2. **Network Communication**: TLS-encrypted TCP connections with persistent connections and proper handshake
3. **Clipboard Monitoring**: Continuous monitoring of clipboard changes
4. **GUI Interface**: Built with Gio UI for cross-platform compatibility

## Implementation Flow

1. **Application Startup**:
   - Generates a unique peer ID for this instance
   - Generates a self-signed TLS certificate for encrypted communication
   - Registers itself on the network using mDNS service discovery on a configurable port (default 50123)
   - Starts a TLS listener to receive clipboard data with proper JSON handshake
   - Begins periodic device discovery every 3 seconds
   - Starts clipboard monitoring every 1 second

2. **Device Discovery**:
   - Uses the `zeroconf` library to discover other ClipSync instances
   - Filters out self from discovered services
   - Attempts to establish persistent TLS connections with proper handshake to discovered devices
   - Uses peer ID as connection key to prevent duplicate connections

3. **Handshake Process**:
   - When connecting to a peer, exchanges JSON handshake with peer ID
   - Uses length-prefix framing to prevent partial frame issues
   - Maintains persistent connections for real-time clipboard synchronization

4. **Clipboard Monitoring**:
   - Continuously checks for clipboard changes
   - When a change is detected, broadcasts it to all connected peers
   - Uses proper synchronization to prevent infinite loops
   - Implements timestamp checking to ignore own changes within 500ms

5. **Connection Management**:
   - Implements graceful shutdown ordering to properly close connections
   - Handles bidirectional communication between peers

## Key Functions

### In `main.go`:

- `NewApp()`: Initializes the application state with a unique peer ID and TLS certificate. It also sets up the GUI theme and initializes the clipboard history and connected devices tracking.

- `startServices()`: Starts all background services including:
  - Service discovery using mDNS
  - TLS listener for incoming connections
  - Periodic device discovery every 3 seconds
  - Clipboard monitoring every 1 second
  This function uses goroutines to run these services concurrently without blocking the main GUI thread.

- `discoverAndSyncDevices()`: Discovers devices on the network using mDNS service discovery. It filters out the local instance to avoid self-connection and establishes TLS connections to other ClipSync instances.

- `sendClipboardToPeers()`: Sends clipboard data to all currently connected peers using the established TLS connections. It uses length-prefix framing to ensure data integrity.

- `shutdown()`: Properly closes all connections when shutting down the application. It follows a specific ordering to ensure clean disconnection from all peers and stops service discovery.

- `Layout()`: Renders the GUI interface using the Gio UI framework. It creates a two-panel layout with clipboard history on the left and connected devices on the right.

### In `modules/network.go`:

- `StartServiceDiscovery()`: Registers the ClipSync service with mDNS so that other instances on the network can discover it. It advertises the service name, type, and port.

- `DiscoverServices()`: Uses zeroconf to browse for other ClipSync services on the local network. It returns a list of discovered services with their IP addresses and ports.

- `ListenForDataTLS()`: Starts a TLS listener on the specified port to accept incoming connections from other ClipSync instances. It handles the TLS handshake and begins receiving clipboard data.

- `handleConnection()`: Processes incoming connections from peers. It handles the initial JSON handshake to identify the peer, then enters a loop to receive clipboard data frames. Each frame is processed by the provided handler function.

- `ConnectToPeerTLS()`: Establishes an outbound TLS connection to a peer ClipSync instance. It performs the handshake exchange to identify itself to the peer.

- `Peer.SendClipboardData()`: Sends clipboard data to a specific peer using length-prefix framing. The data length is sent first as a 4-byte integer, followed by the actual data.

- `Peer.Close()`: Properly closes a peer connection by closing the underlying network connection and setting it to nil.

- `GenerateSelfSignedCert()`: Generates a self-signed TLS certificate for encrypted communication between ClipSync instances. This ensures that clipboard data is transmitted securely.

### In `modules/clipboard.go`:

- `ReadClipboard()`: Reads content from the system clipboard using the atotto/clipboard library. It includes error handling and content validation to ensure only valid text is processed.

- `WriteClipboard()`: Writes content to the system clipboard using the atotto/clipboard library. It includes loop prevention by tracking the last content written.

## Implemented Features

The updated implementation now includes all the features specified in `prompt.txt`:

1. ✅ TLS encryption for secure communication
2. ✅ Proper JSON handshake with peer ID
3. ✅ Persistent connections
4. ✅ Bidirectional synchronization logic
5. ✅ Improved error handling and connection management
6. ✅ Uses length-prefix framing to prevent partial frame issues
7. ✅ Deduplicates IP addresses to prevent duplicate connections
8. ✅ Prevents clipboard race conditions on Windows
9. ✅ UI updates handled properly outside main thread
10. ✅ Configurable port via command line flag
11. ✅ Graceful shutdown ordering
12. ✅ Build tags for cross-platform compatibility
13. ✅ Real-time clipboard synchronization
14. ✅ Instant IP change detection through continuous mDNS browsing

## Solution Implemented

To fix the bidirectional synchronization issue, we made the following changes:

1. **Improved Loop Prevention**: Enhanced the timestamp-based loop prevention mechanism in the data receiving handler to properly track when clipboard updates occur.

2. **Simplified Handler Execution**: Removed the goroutine wrapper around the handler function in `handleConnection` to ensure proper sequential processing of clipboard data.

3. **Better Synchronization**: Ensured that both the clipboard monitoring and data receiving functions properly use mutex locks to prevent race conditions.

These changes ensure that when device1 receives clipboard data from device2, it can still send its own clipboard changes back to device2 without conflicts.