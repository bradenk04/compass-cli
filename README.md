# Compass CLI

A terminal UI for MongoDB, inspired by MongoDB Compass.

> [!WARNING]
> I have been using this tool on my own machine for a while but it is not highly tested.
> There will most likely be bugs I haven't noticed, feel free to open an issue!

![Demo](https://img.shields.io/badge/built%20with-Go-00ADD8?style=flat-square&logo=go)
![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)

## Features

- Browse databases and collections in a collapsible sidebar
- Query documents with filter, sort, limit, and skip inputs
- Insert, edit, and delete documents in an in-terminal JSON editor
- Analyze collection schema with type distribution charts
- Build and execute aggregation pipelines with a live results preview
- View collection indexes
- Persistent connection history (up to 5 recent connections)
- Syntax-highlighted BSON/JSON output

## Installation

### Pre-built binaries (recommended)

Download the latest release for your platform from the [Releases](../../releases) page.

| Platform       | File                                  |
|----------------|---------------------------------------|
| Linux (x86_64) | `compass-linux-amd64`                 |
| macOS (Intel)  | `compass-darwin-amd64`                |
| macOS (Apple Silicon) | `compass-darwin-arm64`         |
| Windows (x86_64) | `compass-windows-amd64.exe`         |

**Linux / macOS:**
```bash
chmod +x compass-<platform>
sudo mv compass-<platform> /usr/local/bin/compass
```

**Windows:**

Move `compass-windows-amd64.exe` somewhere on your `PATH` and rename it to `compass.exe`.

### Build from source

Requires [Go 1.21+](https://go.dev/dl/).

```bash
git clone https://github.com/yourusername/compass-cli.git
cd compass-cli
go build -o compass ./cmd/compass-cli
```

To install directly to your local bin:

```bash
go build -o ~/.local/bin/compass ./cmd/compass-cli
```

## Usage

```bash
compass
```

## Requirements

- A running MongoDB instance (local or remote)
- A terminal with 256-color support
