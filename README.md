# Project Overview
**ClipSync** is a LAN-based clipboard-sharing application that synchronizes clipboards across devices on the same network.  
Architecture: peer-to-peer discovery via **mDNS** (service discovery) and **TCP** (data transfer).

---

# Technical Architecture Assessment

| Component        | Assessment |
|------------------|------------|
| **Discovery**    | **mDNS** (via [zeroconf](https://github.com/grandcat/zeroconf)) is ideal for LAN—zero manual IP entry, seamless UX. |
| **Communication**| **TCP** using net ensures reliable delivery, critical for clipboard integrity. Must handle text, images, files, etc. |
| **Cross-platform GUI** | "Gio UI" gives native-looking UIs on Windows, Linux, macOS from a single Go codebase. |

---

I am using this as a repo and the link is https://github.com/DiamondOsas/ClipSync.git