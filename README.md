# frpc-webui

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

A multi-profile web management UI for [frp](https://github.com/fatedier/frp) client (frpc). Manage multiple frpc proxy configurations through a clean web interface, with persistent storage and Docker support.

## Features

- **Multi-Profile Management** — Create, edit, delete, and switch between multiple frpc proxy profiles
- **Web Admin UI** — Intuitive browser-based interface for managing all proxy configurations
- **CRUD Proxy Operations** — Add, update, and remove TCP/UDP/HTTP proxies dynamically at runtime
- **Persistent Storage** — Profiles are saved to disk and automatically restored on restart
- **Port Conflict Detection** — Real-time detection of port conflicts with existing proxies on the frps server
- **Docker Ready** — Pre-built Docker image for quick deployment
- **API Access** — RESTful API for programmatic profile and proxy management

## Quick Start (Docker)

```bash
docker run -d \
  --name frpc-manager \
  -p 7400:7400 \
  -v frpc_data:/data \
  -e FRPC_WEB_USER=admin \
  -e FRPC_WEB_PASSWORD=your_password \
  huadong7/frpc-webui:latest
```

Then visit `http://localhost:7400` and log in with your configured credentials.

## Build from Source

### Prerequisites

- Go 1.22+
- Node.js 18+ (for web frontend)
- GNU Make

### Build

```bash
# Build both server and client binaries
make build

# Build client binary only (frpc)
make frpc

# Build web assets
make web
```

### Run

```bash
./bin/frpc -c ./conf/frpc.toml
```

Then open `http://127.0.0.1:7400` to access the admin UI.

## Configuration

### Enable Web Admin UI

```toml
# frpc.toml
webServer.addr = "0.0.0.0"
webServer.port = 7400
webServer.user = "admin"
webServer.password = "admin"
```

### Enable Profile Persistence

```toml
[store]
path = "./profiles.json"
```

### Full Client Config Example

See [conf/frpc_full_example.toml](conf/frpc_full_example.toml) for all available options.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/profiles` | List all profiles |
| POST | `/api/profiles` | Create a new profile |
| PUT | `/api/profiles/:name` | Update a profile |
| DELETE | `/api/profiles/:name` | Delete a profile |
| POST | `/api/profiles/:name/start` | Start a profile |
| GET | `/api/profiles/:name/ports/used` | Get used ports on frps server for a profile |

## Acknowledgements

This project is based on [fatedier/frp](https://github.com/fatedier/frp), a fast reverse proxy for exposing local servers behind NAT or firewall to the internet. See the original project for the full frp feature set including TCP/UDP/HTTP/HTTPS proxying, P2P mode, load balancing, encryption, and more.

## License

Apache License 2.0 — see [LICENSE](LICENSE) for details.
