# ClipSync

<p align="center">
  <img src="assets/logo.jpg" alt="ClipSync Logo" width="100">
</p>

 ## Stop emailing yourself links in 2026. Stop Slacking yourself snippets. Stop the friction.

ClipSync synchronizes your clipboard across every device on your local network. It is built for those who value flow state over file transfers. No accounts. No cloud. No latency.

### Why ClipSync

*   **Zero Configuration**: Your devices find each other instantly. No manual IP entry. No handshake. Just sync.
*   **True Privacy**: Your clipboard data never leaves your local network. It moves directly from device A to device B.
*   **Native Performance**: (5MB Ram, 0.1% CPU). It sits in the background and does its job.

### Installation
Download it 

### Contribution
Clone the repository and run the engine.

```bash
git clone https://github.com/DiamondOsas/ClipSync.git
cd ClipSync
go run 
```

### System Requirements

For those running Linux, the following development packages are required to interface with the X11 clipboard system.

**Debian / Ubuntu / Pop!_OS**
```bash
sudo apt install libx11-dev libwayland-dev libxkbcommon-dev libvulkan-dev
```

**Fedora / CentOS / RHEL**
```bash 
sudo dnf install libX11-devel libwayland-dev libxkbcommon-dev vulkan-headers 
```

**Arch Linux / Manjaro**
```bash
sudo pacman -S libx11 libwayland-dev libxkbcommon-dev libvulkan-dev
```

### Development

ClipSync is currently in its early stages of development. If you experience any bug during use open an ISSUE. 


