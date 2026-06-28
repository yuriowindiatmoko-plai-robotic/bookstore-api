# Bookstore REST API

A REST API for a bookstore, built from scratch with Go + Gin + GORM on top of PostgreSQL 16 running in **Podman** (no Docker).

_REST API untuk toko buku, dibangun dari nol memakai Go + Gin + GORM di atas PostgreSQL 16 yang berjalan di **Podman** (tanpa Docker)._

---

## Tech Stack · _Tumpukan Teknologi_

| Layer | Choice | Version |
|---|---|---|
| Language | Go | 1.25.0 |
| HTTP framework | Gin | 1.12.0 |
| ORM | GORM + `gorm.io/driver/postgres` (pgx) | 1.31.2 / 1.6.0 |
| Auth | `golang-jwt/v5` (planned), bcrypt (`x/crypto`) | v5.3.1 / v0.48.0 |
| Env loader | `joho/godotenv` | 1.5.1 |
| Database | PostgreSQL | 16 |
| Container runtime | Podman + `docker-compose` provider | 6.0.0 / 5.2.0 |
| Host | macOS (Apple Silicon) | — |

---

## Project Structure · _Struktur Proyek_

```
bookstore-api/
├── main.go                 # Entry point: connect, migrate, Gin setup, /health, serve
├── config/
│   └── database.go         # LoadEnv() + Connect() — retry, pool tuning, Gin-routed logs
├── models/
│   ├── book.go             # Book GORM model
│   └── user.go             # User GORM model + bcrypt HashPassword/CheckPassword
├── cmd/
│   └── dbcheck/
│       └── main.go         # Standalone DB connectivity checker
├── docker-compose.yml      # PostgreSQL 16 service (Podman-compatible)
├── .env                    # Environment variables (gitignored)
├── .gitignore
├── go.mod / go.sum
└── AGENTS.md               # Project conventions
```

---

## Prerequisites & Setup (macOS / Apple Silicon) · _Prasyarat_

This is the part that cost the most sweat — Podman on Apple Silicon needs the `libkrun`/`krunkit` provider, which has several non-obvious traps. _Bagian ini yang paling menyusahkan — Podman di Apple Silicon butuh provider `libkrun`/`krunkit` yang punya beberapa jebakan tak terduga._

### 1. Install Podman, the compose provider, and Go

```bash
brew install podman docker-compose
brew upgrade go          # needs Go >= 1.23 (we run 1.25)
```

### 2. Provide the `krunkit` binary (not in Homebrew)

Podman 6 defaults to the `libkrun` provider, which calls `krunkit`. Since it's not packaged in Homebrew, pull it from GitHub and ad-hoc sign it:

```bash
curl -LO https://github.com/libkrun/krunkit/releases/download/v1.2.2/krunkit-podman-unsigned-1.2.2.tgz
tar -xzf krunkit-podman-unsigned-1.2.2.tgz
# Place binary + dylibs in the Homebrew prefix (matches krunkit's @rpath)
install -m 0755 bin/krunkit      /opt/homebrew/bin/krunkit.real
install -m 0644 lib/*.dylib      /opt/homebrew/lib/
mkdir -p /opt/homebrew/share/krunkit
install -m 0644 share/krunkit/KRUN_EFI.silent.fd /opt/homebrew/share/krunkit/
```

### 3. Sign krunkit with **only** the hypervisor entitlement

Gotcha #1: on Apple Silicon, unsigned krunkit gets SIGKILL'd (exit 137). Gotcha #2: ad-hoc signing with the private `com.apple.vm.*` entitlements **also** gets SIGKILL'd — only the public `com.apple.security.hypervisor` entitlement is allowed for ad-hoc signatures.

```bash
cat > /tmp/hv.plist <<'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>com.apple.security.hypervisor</key><true/>
</dict></plist>
EOF
codesign -s - --force --entitlements /tmp/hv.plist /opt/homebrew/bin/krunkit.real
```

### 4. Strip the `--nested` flag via a wrapper

Gotcha #3: Podman passes `--nested` to krunkit unconditionally, but enabling nested virtualization on the host gets krunkit killed. Nested virt is irrelevant for running containers, so wrap it out:

```bash
cat > /opt/homebrew/bin/krunkit <<'EOF'
#!/bin/zsh
args=()
for a in "$@"; do [[ "$a" == "--nested" ]] && continue; args+=("$a"); done
exec /opt/homebrew/bin/krunkit.real "${args[@]}"
EOF
chmod +x /opt/homebrew/bin/krunkit
```

### 5. Fix the Docker credential helper

Gotcha #4: `~/.docker/config.json` references Docker Desktop's `docker-credential-desktop`. Remove the `credsStore` line so public image pulls work:

```json
{ "auths": {}, "currentContext": "default" }
```

### 6. Create and start the Podman machine

```bash
podman machine init
podman machine start
```

Gotcha #5: if `podman compose` ever errors with a missing `…-api.sock`, just restart the machine — `podman machine stop && podman machine start` recreates the gvproxy API-forwarding socket.

---

## Quick Start · _Mulai Cepat_

```bash
# 1. Boot the Podman VM (once per reboot)
podman machine start

# 2. Start PostgreSQL
podman compose up -d

# 3. Run the API
go run .

# 4. Smoke test
curl http://localhost:8080/health
# {"data":{"status":"ok"}}
```

---

## Environment Variables · _Variabel Lingkungan_

Defined in `.env` (gitignored). _Didefinisikan di `.env` (diabaikan oleh git)._

| Variable | Default | Description |
|---|---|---|
| `DB_HOST` | `localhost` | Postgres host |
| `DB_PORT` | `5432` | Postgres port |
| `DB_USER` | `bookstore` | DB user |
| `DB_PASSWORD` | `bookstore_secret` | DB password |
| `DB_NAME` | `bookstore_db` | DB name |
| `DB_SSLMODE` | `disable` | SSL mode |
| `DB_TIMEZONE` | `UTC` | DB timezone |
| `DB_MAX_OPEN_CONNS` | `25` | Pool: max open connections |
| `DB_MAX_IDLE_CONNS` | `5` | Pool: max idle connections |
| `DB_CONN_MAX_LIFETIME` | `30m` | Pool: connection max lifetime |
| `DB_CONN_MAX_IDLE_TIME` | `5m` | Pool: connection max idle time |
| `JWT_SECRET` | — | JWT signing secret (change in production!) |
| `JWT_EXPIRY` | `24h` | JWT token lifetime (planned) |
| `PORT` | `8080` | Server port |
| `GIN_MODE` | `debug` | `debug` / `release` / `test` |

---

## API Endpoints · _Endpoint API_

| Method | Path | Status | Description |
|---|---|---|---|
| `GET` | `/health` | ✅ done | Health check → `{"data":{"status":"ok"}}` |
| _CRUD_ | _`/api/v1/books`_ | 🔜 planned | Book resource routes (grouped under `/api/v1`) |
| _Auth_ | _`/api/v1/auth/*`_ | 🔜 planned | Register/login, JWT issuance |

Response envelope follows AGENTS.md: success → `{"data": {...}}`, errors → `{"error": "message"}`.

---

## Data Models · _Model Data_

### `Book` (table: `books`)

| Field | Go type | GORM tag | JSON |
|---|---|---|---|
| ID | `uint` | `primaryKey` | `id` |
| Title | `string` | `type:varchar(255);not null` | `title` |
| Author | `string` | `type:varchar(255);not null` | `author` |
| ISBN | `string` | `type:varchar(57);uniqueIndex` | `isbn` |
| Price | `float64` | `type:numeric(10,2)` | `price` |
| Stock | `int` | `default:0` | `stock` |
| Description | `*string` | `type:text` | `description` |
| CreatedAt | `time.Time` | auto | `created_at` |
| UpdatedAt | `time.Time` | auto | `updated_at` |

`Description` is a pointer (`*string`) for true SQL `NULL` semantics. _`Description` memakai pointer agar bisa menyimpan SQL `NULL` sebenarnya._

### `User` (table: `users`)

| Field | Go type | GORM tag | JSON |
|---|---|---|---|
| ID | `uint` | `primaryKey` | `id` |
| Name | `string` | `type:varchar(255);not null` | `name` |
| Email | `string` | `type:varchar(255);uniqueIndex;not null` | `email` |
| Password | `string` | `type:varchar(255);not null` | **`-`** (never serialized) |
| CreatedAt | `time.Time` | auto | `created_at` |
| UpdatedAt | `time.Time` | auto | `updated_at` |

Methods (pointer receivers):
- `HashPassword(password string) error` — bcrypt at `DefaultCost` (10), stores hash in `Password`.
- `CheckPassword(password string) bool` — wraps `bcrypt.CompareHashAndPassword`.

`Password` has `json:"-"` so it is **never** returned in any response. _`Password` diberi `json:"-"` agar tidak pernah dikembalikan di response._

---

## Database Connection · _Koneksi Database_

`config.Connect()` (`config/database.go`):

- **Loads `.env`** via godotenv (falls back to process env if missing).
- **Retry:** 3 attempts, fixed **5s** delay between attempts (no trailing delay).
- **Pool tuning** (defaults above, overridable via env).
- **Logging** routed through Gin's writers (`gin.DefaultWriter` / `gin.DefaultErrorWriter`); GORM's own logger is wired to `gin.DefaultWriter` — no `fmt.Println`, per AGENTS.md.
- Returns `*gorm.DB`; on final failure returns a wrapped error (raw DB errors stop at `main.go`, never reaching HTTP clients).

### Memory cost per Postgres connection · _Beban RAM per koneksi_

Measured empirically on this setup (`shared_buffers=128MB`, `work_mem=4MB`, `max_connections=100`):

| Connection state | RSS (`ps`) | Private (RssAnon) |
|---|---|---|
| Idle Go-app backend (queries run) | ~25.5 MB | ~9.3 MB |
| Fresh idle backend | ~14.9 MB | ~2.9 MB |

**Rule of thumb:** budget ~5–10 MB of *exclusive* RAM per connection. The ~15 MB RSS figure counts shared memory (`shared_buffers` + postgres code/libs) that is amortized across all backends — only `RssAnon` is truly per-connection. _Anggaran ~5–10 MB RAM eksklusif per koneksi. Angka ~15 MB RSS menghitung memori shared yang dibagi semua backend — hanya `RssAnon` yang benar-benar per-koneksi._

Implication: with `MaxIdleConns=5`, idle steady-state is ~15 MB private; at `MaxOpenConns=25` under load, worst case ~325 MB private. Beyond ~50 concurrent connections, add **PgBouncer** rather than spawning more backends.

---

## Development Commands · _Perintah Pengembangan_

```bash
# Build
go build -o bookstore-api .

# Run
go run .                       # or: ./bookstore-api

# Check DB connectivity
go run ./cmd/dbcheck

# Lint / format
go vet ./...
gofmt -l .

# PostgreSQL lifecycle (Podman)
podman compose up -d           # start postgres
podman compose ps              # status
podman compose logs -f postgres
podman compose down            # stop

# Inspect Postgres directly
podman exec bookstore-postgres psql -U bookstore -d bookstore_db
```

---

## Roadmap · _Peta Jalan_

**Done · _Selesai_**
- Project scaffolding (Go module, deps)
- PostgreSQL 16 on Podman + `docker-compose.yml` + `.env`
- `config/` — env loading, GORM connect with retry + pool tuning + Gin logging
- `models/` — `Book` and `User` (bcrypt)
- `main.go` — connect, auto-migrate, Gin (logger + recovery), `GET /health`, serve
- DB connectivity checker (`cmd/dbcheck`)

**Next · _Berikutnya_**
- `handlers/` — Gin handlers for Book CRUD + User auth
- `routes/` — route registration grouped under `/api/v1`
- `middleware/` — JWT auth middleware (`golang-jwt/v5`)
- Input validation (Gin binding)
- Pagination per AGENTS.md (`?page=1&limit=10`)
- Tests

---

## Notes · _Catatan_

- `.env` is gitignored — never commit secrets. The Postgres credentials in `docker-compose.yml` are dev-only placeholders.
- Gin startup warnings (`debug` mode, "trusted all proxies") are expected for dev — set `GIN_MODE=release` and configure trusted proxies for production.
- This project follows the conventions in [`AGENTS.md`](./AGENTS.md).

Built end-to-end in a single session — from an empty directory to a running, migrating, health-checked API. _Dibangun lengkap dalam satu sesi — dari direktori kosong menjadi API yang berjalan, migrasi, dan health-checked._
