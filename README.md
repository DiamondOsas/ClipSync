# ✂️ ClipSync — Clipboard Sync That Actually Works

<p align="center">
  <img src="assets/logo.jpg" alt="ClipSync logo" width="120">
</p>

<p align="center">
  <b>Copy on your laptop. Paste on your desktop. Done.</b><br>
  <sub>No cloud. No account. No friction. Just your clipboard, everywhere on your network.</sub>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Built with Go">
  <img src="https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey?style=for-the-badge" alt="Windows, macOS, Linux">
  <img src="https://img.shields.io/badge/Network-LAN%20Only%2C%20No%20Cloud-green?style=for-the-badge" alt="Local network only">
  <a href="https://diamondosas.github.io/clipsync/">
    <img src="https://img.shields.io/badge/Download-Now-blue?style=for-the-badge&logo=appveyor" alt="Download ClipSync">
  </a>
</p>

---

**ClipSync** is an open-source clipboard tool that syncs across every device on your local network — Windows, macOS, Linux — instantly. Built in Go. Runs silent. Uses barely any RAM. No servers, no accounts, no nonsense.

> Emailing yourself links in 2026 is embarrassing. Stop it.

---

<h3 align="center">🎬 Demo</h3>

<p align="center">
  <!-- Drop your GIF here: assets/demo.gif -->
  <img src="assets/demo.gif" alt="ClipSync demo" width="700">
</p>

---

### 🚨 Why It Matters

<table>
<tr>
<td width="33%" align="center" valign="top">⚡<br><b>Instant sync</b><br><sub>Copy on one machine,<br>paste on another — fast</sub></td>
<td width="33%" align="center" valign="top">🔒<br><b>100% private</b><br><sub>Data never leaves<br>your local network</sub></td>
<td width="33%" align="center" valign="top">🆓<br><b>Free forever</b><br><sub>No subscriptions.<br>No premium tiers.</sub></td>
</tr>
<tr>
<td align="center" valign="top">🌐<br><b>Works everywhere</b><br><sub>Windows · macOS · Linux<br>X11 and Wayland</sub></td>
<td align="center" valign="top">🔌<br><b>Zero setup</b><br><sub>Devices find each other<br>automatically on LAN</sub></td>
<td align="center" valign="top">🪶<br><b>Tiny footprint</b><br><sub>~5 MB RAM · ~0.1% CPU<br>runs silently in background</sub></td>
</tr>
</table>

---

## 📥 Download ClipSync

Get the latest binary for your OS — run it, done.

**[⬇️ Download ClipSync — Windows, macOS, Linux](https://diamondosas.github.io/clipsync/)**

---

## 🆚 ClipSync vs. Everything Else

Most clipboard tools do too much. Mouse sharing, file transfer, notification mirroring — nobody asked for that. ClipSync does one thing: clipboard sync. Lean. Fast. Free.

| Tool | Free | No Cloud | Cross-Platform | Open Source |
|---|---|---|---|---|
| **ClipSync** | ✅ | ✅ | ✅ | ✅ |
| KDE Connect | ✅ | ✅ | Partial | ✅ |
| Synergy | ❌ | ✅ | ✅ | Partial |
| ShareMouse | ❌ | ✅ | ✅ | ❌ |
| Pushbullet | Partial | ❌ | ✅ | ❌ |
| Apple Handoff | ✅ | ❌ | Apple only | ❌ |

---

## 🚀 Get It Running in 60 Seconds

### Option 1: Download the Binary

Go to **[diamondosas.github.io/clipsync](https://diamondosas.github.io/clipsync/)**, grab the file for your OS, run it. That's it. ClipSync starts listening on your network immediately.

### Option 2: Build From Source

```bash
# Clone the repo
git clone https://github.com/DiamondOsas/ClipSync.git
cd ClipSync

# Run it
go run main.go
```

---

## ⚙️ Linux Setup — One Extra Step

On Linux, ClipSync needs access to the X11 or Wayland clipboard layer. Install the right libraries for your distro first:

**Debian / Ubuntu / Pop!_OS / Mint**
```bash
sudo apt install libx11-dev libwayland-dev libxkbcommon-dev libvulkan-dev
```

**Fedora / CentOS / RHEL / AlmaLinux**
```bash
sudo dnf install libX11-devel libwayland-dev libxkbcommon-dev vulkan-headers
```

**Arch / Manjaro / EndeavourOS**
```bash
sudo pacman -S libx11 libwayland-dev libxkbcommon-dev libvulkan-dev
```

---

## 🔍 How It Works

ClipSync finds other ClipSync devices on your network automatically — no IPs, no pairing screens, no config files. When you copy something, it broadcasts to all connected devices over your LAN. Near-instant. Never leaves your network.

**Use it for:**
- Sync clipboard between your Windows PC and MacBook on the same Wi-Fi
- Copy terminal output on a Linux server, paste it locally
- Share a URL across machines in your home office without opening Slack
- Move code snippets between desktop and laptop, fast

---

## 🐛 Found a Bug? Want a Feature?

ClipSync is actively built. If something breaks or you have an idea:

👉 **[Open an issue on GitHub](https://github.com/DiamondOsas/ClipSync/issues)**

---

## ⭐ Star History

If ClipSync saves you time, star it. Helps other people find it.

<!-- <picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=diamondosas/clipsync&type=date&theme=dark&legend=top-left" />
  <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=diamondosas/clipsync&type=date&legend=top-left" />
  <img alt="ClipSync GitHub star history" src="https://api.star-history.com/chart?repos=diamondosas/clipsync&type=date&legend=top-left" />
</picture> -->

---

## 🏷️ Keywords

`clipboard sync` · `cross-device clipboard` · `local network clipboard` · `LAN clipboard manager` · `copy paste between computers` · `clipboard sharing tool` · `sync clipboard Windows macOS Linux` · `offline clipboard sync` · `no cloud clipboard` · `KDE Connect alternative` · `developer productivity tool` · `Go clipboard tool`
