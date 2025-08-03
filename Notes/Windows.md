# Cross-Clip | Windows Edition  
> One-file, dark-themed clipboard daemon that syncs text, images, and files across Windows, Linux, Android, and (soon) iOS.  
> **Core rule:** zero cloud, zero sign-up, zero bloat.

---

## 1. Elevator Pitch
A 6 MB portable executable (`cross-clip.exe`) you drop anywhere.  
Run once → it sits in the tray with a frosted-glass, dark-mode UI → copies propagate to every paired device in < 200 ms over **Wi-Fi Direct, LAN, or Bluetooth**.

---

## 2. Core Functionality (Windows MVP)
| Feature | Status | Notes |
|---|---|---|
| **Listen** | ✅ | Hook `WM_CLIPBOARDUPDATE`; capture text, RTF, PNG, 32-bit bitmap, file-drop (`CF_HDROP`). |
| **Send** | ✅ | Push to any device via chosen transport (see §3). |
| **Receive** | ✅ | Auto-merge incoming clip into local clipboard; optional toast + sound. |
| **History** | ⚡ | `Ctrl + Shift + V` opens a dark, searchable pop-up with 100 last clips. |
| **Security** | ✅ | NaCl secretbox + per-device pre-shared key (QR code exchange). |
| **Offline** | ✅ | Works with zero internet; only local network required. |

---

## 3. Transport Layers (plug-and-play)

| Mode | Discovery | Bandwidth | Setup Steps |
|---|---|---|---|
| **LAN mDNS** | `_crossclip._tcp` broadcast | 1 Gb/s | Works out-of-box on home/office Wi-Fi. |
| **Wi-Fi Direct** | Windows “Project to this PC” API | 250 Mb/s | Pair once via PIN or QR. |
| **Bluetooth LE** | GATT service `a4e649f2-abe8-11ed-afa1-0242ac120002` | 2 Mb/s | Toggle in tray menu; auto-pair with nearby devices. |
| **Manual IP** | Text box | — | For headless servers or VPN. |

> All modes run concurrently; the daemon picks the fastest path automatically.

---

## 4. Dark Theme Design Spec
```
background  : #0d1117
surface     : #161b22
border      : #30363d
accent      : #58a6ff (cyan-blue)
text        : #c9d1d9
selection   : rgba(88,166,255,0.25)
```

UI elements:
- **Tray icon**: monochrome moon glyph, turns cyan when syncing.  
- **History window**: rounded 8 px, acrylic blur (Windows 11 `Mica`), 80 % opacity.  
- **QR Pairing**: full-screen overlay with dark backdrop and cyan corner brackets.

---

## 5. Quick-Start (Windows)
1. Download `cross-clip-v0.1.0.zip` (6.3 MB).  
2. Unzip → run `cross-clip.exe`.  
3. Tray icon → “Pair new device” → choose transport → scan QR or enter PIN.  
4. Copy anything; see it appear on your Android (or another PC) instantly.

---

## 6. Build from Source (if you want)
```bash
git clone https://github.com/you/cross-clip
cd cross-clip
go mod tidy
go build -ldflags "-s -w -H=windowsgui" -o cross-clip.exe ./cmd/win
```
> Requires Go 1.22+ (download ~150 MB once).

---

| Layer                        | Technology                                                              | Rationale                                                                                                                 |
| ---------------------------- | ----------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| **Language & Runtime**       | Go 1.22                                                                 | Single static binary (~6 MB), no runtime install, cross-compile to Windows with `-H=windowsgui` for headless tray daemon. |
| **Clipboard Hook**           | Win32 API (`AddClipboardFormatListener`) via `golang.org/x/sys/windows` | Native, zero-dependency listener for text, bitmap, RTF, file-drop (`CF_HDROP`).                                           |
| **Transport (LAN)**          | mDNS + TCP                                                              | Zero-config discovery via `_crossclip._tcp` multicast; raw TCP for payload.                                               |
| **Transport (Wi-Fi Direct)** | Windows UWP `Windows.Devices.WiFiDirect`                                | Leverages built-in “Project to this PC” stack; no external driver.                                                        |
| **Transport (Bluetooth)**    | WinRT GATT Server & Client                                              | BLE service UUID `a4e649f2-abe8-11ed-afa1-0242ac120002`; 2 Mb/s, 15 m range.                                              |
| **Encryption**               | libsodium/NaCl (`secretbox`)                                            | 256-bit XSalsa20-Poly1305, per-device pre-shared key exchanged via QR code.                                               |
| **Local Storage**            | BadgerDB (embedded key-value)                                           | Pure-Go, disk-backed clip history (<10 MB).                                                                               |
| **UI**                       | Fyne (Go) or Win32 Tray + Wails                                         | Dark acrylic/mica window (#0d1117, #58a6ff accent) with frosted-glass pop-up history.                                     |
| **Build & Distribution**     | `go build -ldflags "-s -w -H=windowsgui"`                               | Single portable EXE; no installer.                                                                                        |
