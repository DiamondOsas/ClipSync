# ✂️ ClipSync — Free Cross-Device Clipboard Sync for Local Networks

<p align="center">
  <img src="assets/logo.jpg" alt="ClipSync - clipboard sync tool for Windows, macOS, and Linux" width="120">
</p>

<p align="center">
  <b>Stop emailing yourself links in 2026. Stop Slacking yourself snippets. Stop the friction.</b><br>
  <sub>Instant clipboard sharing across all your devices — no cloud, no account, no latency.</sub>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Built with Go">
  <img src="https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey?style=for-the-badge" alt="Supported platforms: Windows, macOS, Linux">
  <img src="https://img.shields.io/badge/Network-LAN%20Only%2C%20No%20Cloud-green?style=for-the-badge" alt="Local network only, no cloud">
  <a href="https://diamondosas.github.io/clipsync/">
    <img src="https://img.shields.io/badge/Download-Now-blue?style=for-the-badge&logo=appveyor" alt="Download ClipSync">
  </a>
</p>

---

**ClipSync** is an open-source, local-network clipboard manager that instantly synchronizes your clipboard across every device — Windows, macOS, and Linux — without an internet connection, account, or cloud service. Built in Go for native performance, it is designed for developers, power users, and anyone who works across multiple machines every day.

> Copy on your laptop. Paste on your desktop. It just works.

---

## 📥 Download ClipSync

Get the latest compiled binary for your operating system:

**[⬇️ Download ClipSync — Windows, macOS, Linux](https://diamondosas.github.io/clipsync/)**

---

## ✨ Features — Why ClipSync?

| Feature | ClipSync |
|---|---|
| Works on LAN only (no internet required) | ✅ |
| Zero accounts or sign-up | ✅ |
| Open source | ✅ |
| Auto device discovery | ✅ |
| Windows + macOS + Linux | ✅ |
| RAM usage | ~5 MB |
| CPU usage | ~0.1% |

- 🔌 **Zero-Config Device Discovery** — Devices on the same local network find each other automatically. No IP addresses to set. No pairing screens.
- 🔒 **100% Private, Offline Clipboard Sync** — Your clipboard data never touches a server. It travels directly from device to device over your LAN.
- ⚡ **Native Performance via Go** — ClipSync is tiny (~5 MB RAM, 0.1% CPU) and runs silently in the background.
- 🌐 **Cross-Platform** — Full support for Windows, macOS, and Linux (X11 & Wayland).
- 🆓 **Free and Open Source** — No subscriptions. No premium tiers. Clone it, build it, ship it.

---

## 🆚 ClipSync vs. Other Clipboard Sync Tools

Looking for a **KDE Connect alternative**? A **Synergy alternative**? A **ShareMouse alternative** that doesn't cost money? ClipSync is purpose-built for clipboard synchronization only — which means it is leaner, simpler, and faster than tools that bundle mouse/keyboard sharing, file transfer, and notification mirroring.

| Tool | Free | No Cloud | Cross-Platform | Open Source |
|---|---|---|---|---|
| **ClipSync** | ✅ | ✅ | ✅ | ✅ |
| KDE Connect | ✅ | ✅ | Partial | ✅ |
| Synergy | ❌ | ✅ | ✅ | Partial |
| ShareMouse | ❌ | ✅ | ✅ | ❌ |
| Pushbullet | Partial | ❌ | ✅ | ❌ |
| Apple Handoff | ✅ | ❌ | Apple only | ❌ |

---

## 🚀 Quick Start — Getting ClipSync Running

### Option 1: Download a Pre-Built Binary

Head to the [ClipSync download page](https://diamondosas.github.io/clipsync/) and grab the binary for your OS. Run it, and ClipSync starts listening on your local network immediately.

### Option 2: Build From Source

ClipSync is built with Go. If you want to compile it yourself or contribute:

```bash
# Clone the ClipSync repository
git clone https://github.com/DiamondOsas/ClipSync.git
cd ClipSync

# Run ClipSync locally
go run main.go
```

---

## ⚙️ Linux Setup — System Requirements for Clipboard Access

On Linux, ClipSync needs low-level access to the X11 or Wayland clipboard subsystem. Install the appropriate development libraries for your distro before building from source.

**Debian / Ubuntu / Pop!_OS / Linux Mint**
```bash
sudo apt install libx11-dev libwayland-dev libxkbcommon-dev libvulkan-dev
```

**Fedora / CentOS / RHEL / AlmaLinux**
```bash
sudo dnf install libX11-devel libwayland-dev libxkbcommon-dev vulkan-headers
```

**Arch Linux / Manjaro / EndeavourOS**
```bash
sudo pacman -S libx11 libwayland-dev libxkbcommon-dev libvulkan-dev
```

---

## 🔍 How ClipSync Works

ClipSync uses **LAN-based peer discovery** to find other ClipSync instances on your network without any manual configuration. When you copy something, ClipSync broadcasts the clipboard payload to connected peers — text, URLs, code snippets, anything your OS clipboard supports. The sync is near-instantaneous and never exits your local network.

**Use cases:**
- Sync clipboard between Windows PC and MacBook on the same Wi-Fi
- Share code snippets between your desktop and laptop without Slack or email
- Copy a URL on one machine, paste it on another in your home office or studio
- Developer workflow: copy terminal output on a Linux server, paste it locally

---

## 🐛 Issues, Bugs & Feature Requests

ClipSync is actively developed and welcomes community feedback. If you run into a bug or have an idea for a new feature:

👉 **[Open an issue on GitHub](https://github.com/DiamondOsas/ClipSync/issues)**

---

## ⭐ Star History

If ClipSync saves you time, give it a star — it helps other developers find it.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=diamondosas/clipsync&type=date&theme=dark&legend=top-left" />
  <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=diamondosas/clipsync&type=date&legend=top-left" />
  <img alt="ClipSync GitHub star history chart" src="https://api.star-history.com/chart?repos=diamondosas/clipsync&type=date&legend=top-left" />
</picture>

---

## 🏷️ Keywords

`clipboard sync` · `cross-device clipboard` · `local network clipboard` · `LAN clipboard manager` · `copy paste between computers` · `clipboard sharing tool` · `sync clipboard Windows macOS Linux` · `offline clipboard sync` · `no cloud clipboard` · `KDE Connect alternative` · `developer productivity tool` · `Go clipboard tool`
