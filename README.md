# ✂️ ClipSync

<p align="center">
  <img src="assets/logo.jpg" alt="ClipSync Logo" width="120">
</p>

<p align="center">
  <b>Stop emailing yourself links in 2026. Stop Slacking yourself snippets. Stop the friction.</b>
</p>

<p align="center">
  <!--<a href="YOUR_SOURCEFORGE_LINK_HERE">
    <img src="https://img.shields.io/badge/Download-SourceForge-orange?style=for-the-badge&logo=sourceforge" alt="Download on SourceForge">
  </a>-->
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Made with Go">
  <!--<img src="https://img.shields.io/github/downloads/DiamondOsas/ClipSync/total?style=for-the-badge">  -->
  <a href="https://diamondosas.github.io/clipsync/">
      <img src="https://img.shields.io/badge/Download-Now-blue?style=for-the-badge&logo=appveyor" alt="Download">
    </a>

</p>


ClipSync synchronizes your clipboard across every device on your local network. It is built for developers and power users who value flow state over file transfers. No accounts. No cloud. No latency.

<!--![Demo GIF Placeholder](assets/demo.gif) *(Tip: Record a quick 5-second GIF showing you copying text on one screen and pasting on another!)* -->

---

### 📥 Download
You can download the latest compiled version directly from SourceForge:
**[Download ClipSync Here](YOUR_SOURCEFORGE_LINK_HERE)**

## ✨ Why ClipSync?

* 🔌 **Zero Configuration:** Your devices find each other instantly. No manual IP entry. No handshake. Just sync.
* 🔒 **True Privacy:** Your clipboard data never leaves your local network. It moves directly from device A to device B.
* ⚡ **Native Performance:** Incredibly lightweight (~5MB RAM, 0.1% CPU). It sits quietly in the background and does its job.

## 🚀 Getting Started



### 💻 Development & Contribution
Want to build it from the source or contribute? Clone the repository and run the engine locally.

```bash
git clone [https://github.com/DiamondOsas/ClipSync.git](https://github.com/DiamondOsas/ClipSync.git)
cd ClipSync
go run main.go 
```

## ⚙️ Linux System Requirements

For those running Linux, the following development packages are required to interface with the X11/Wayland clipboard system.

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

## 🐛 Issues & Support
ClipSync is currently in its early stages of development. If you experience any bugs or have feature requests, please [open an issue](https://github.com/DiamondOsas/ClipSync/issues).

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=diamondosas/clipsync&type=date&theme=dark&legend=top-left" />
  <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=diamondosas/clipsync&type=date&legend=top-left" />
  <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=diamondosas/clipsync&type=date&legend=top-left" />
</picture>



