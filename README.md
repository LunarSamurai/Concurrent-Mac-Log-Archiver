# =============================================
#       .logarchive Structure (macOS)
# =============================================

This archive is built to be compatible with macOS `log collect` and Console.app.

# Top-Level Layout

-----------------
MyCase.logarchive/ ..
├── uuidtext/             ← Binary `.tracev3` logs from /private/var/db/uuidtext  
├── system_logs/          ← Traditional system logs from /private/var/log  
├── diagnostics/          ← Crash logs, spin reports from /private/var/db/diagnostics  
├── user_logs/            ← App logs from /Users/<user>/Library/Logs  
├── timesync/             ← Time sync data from /private/var/db/timesync  
├── live/                 ← Live buffer streams (may be empty in deadbox)  
├── LogStoreMetadata/     ← Logd persistence metadata from /private/var/db/logd  
├── network/              ← DiagnosticMessages from /private/var/log  
├── Info.plist            ← Required metadata for macOS log recognition  
└── [Optional folders]    ← Persist/, Metadata/, etc. (optional for basic parsing)

# Usage on macOS

-----------------
```$ log show --style syslog --info --predicate 'eventMessage contains "Safari"' --archive /path/to/MyCase.logarchive```

This will show all events captured inside the archive using macOS-native tools.

Built using: log (Go-based log archivecollection)

-----------------

# Script Functionality

This Go tool mimics the behavior of `log collect` on macOS, and is built for **deadbox forensics** — where you cannot run commands live on the machine.

-----------------

## Modes

### 1. `--input-dir` Mode (Recommended for Deadbox Use)

You specify a folder that contains a full copy of `/private` and `/Users` from the target macOS system.  
The script will:

- Copy critical log folders from the given disk structure
- Preserve directory layout and metadata
- Create a fully structured `.logarchive` compatible with macOS
- Generate `Info.plist` with timestamp and basic metadata

### 2. Interactive Mode (Manual Entry)

If no `--input-dir` is provided, the script will prompt you for each required directory path interactively.  
Use this if you've selectively extracted individual folders from an image or volume.

---

# Files Collected (Details)

| Archive Subdir      | Source Path on macOS                        | Purpose                                                |
|---------------------|---------------------------------------------|--------------------------------------------------------|
| uuidtext/           | /private/var/db/uuidtext                    | Unified logging in .tracev3 format                     |
| system_logs/        | /private/var/log                            | Syslogs, ASL logs, kernel & boot messages              |
| diagnostics/        | /private/var/db/diagnostics                 | Crash reports, hangs, diagnostics                      |
| user_logs/          | /Users/username/Library/Logs                | App logs and custom user-space logs                    |
| timesync/           | /private/var/db/timesync                    | NTP sync data for time correlation                     |
| live/               | /private/var/db/logd/streams                | Live memory streams (usually stale in deadbox)         |
| LogStoreMetadata/   | /private/var/db/logd                        | Internal metadata for logd                             |
| network/            | /private/var/log/DiagnosticMessages         | Network diagnostics from wireless & stack subsystems   |

-----------------

# Example: Viewing Logs on macOS

Once the `.logarchive` is created, move it to a macOS system and use:
You may also load it into Console.app for GUI-based review.

-----------------

# Requirements
- Go 1.18+
- Extracted or mounted macOS filesystem (e.g., from .dmg, .img, .E01, etc.)
- Sufficient read permissions to copied directories

-----------------

# Notes
- This script does not parse .tracev3 — it focuses on collection and packaging only.
- Works in offline environments with full folder access.
- .logarchive is portable across macOS systems running Console.app or the log CLI.
