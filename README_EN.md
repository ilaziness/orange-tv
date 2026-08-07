# 小橘TV

> 中文版主文档：[README.md](README.md) · Technical docs: [docs/README.md](docs/README.md)

An out-of-the-box video streaming site system for building an online video website for viewers. It supports video collection, content management, playback and backend operations, with a responsive front-end for phones, tablets and desktops.

## What can this system do?

A viewing website for regular audiences. Visitors can:

- See recommended videos, category entries and the latest content on the **home page**
- Browse videos by **category** (movies, TV series, variety shows, anime, etc.)
- Filter and sort videos by **region, year, language**
- Open a **video detail page** to see poster, synopsis, director, cast, rating, play sources and episode list
- Watch videos smoothly on the **play page** — switch play sources, select episodes, fast-forward/rewind, fullscreen
- **Search** to quickly find videos they want to watch
- Get a good browsing experience on phones, tablets and desktops (responsive)
- Toggle light/dark theme

## Who uses it?

The system is split into two independent parts for different people:

| Part | Users | Purpose |
| ------ | ------ | ------ |
| **Client** (front-end site) | Regular viewers | Browse, search, watch videos |
| **Admin backend** | Site owners / operators / editors | Manage content, configure collection, manage users, configure site |

## What the admin backend can do

- **Home dashboard**: see today's new videos, total videos, collection task status and other operational data at a glance
- **Content management**
  - Category management: multi-level category CRUD, sorting, enable/disable
  - Video management: create, edit, delete, publish/unpublish, bulk operations
  - Live management: TV live channel CRUD, categories and stream URLs
  - Director / actor / tag management
  - Play source management: global play sources and episode play links
  - Banner management: home carousel configuration
  - Data collection: configure collection sources, scheduled collection, category mapping, play source binding, view collection logs
- **User management**: admin accounts, user groups (roles), regular users, login logs
- **System settings**: site name, logo, copyright, filing number, SEO keywords, API config, system logs

## Data collection (key feature)

If you don't want to enter videos one by one manually, use the collection feature to pull video data from other sites automatically:

- Supports both **default format** (system custom) and **Apple CMS format** collection sources
- Supports **scheduled collection** (configured via cron expression; empty means no auto-collection)
- Also supports **manual one-click trigger** to collect from a source immediately
- Supports **field mapping** and **category mapping** to align external data with your site structure
- Collected play links are automatically routed to the play source you bound
- Full **collection logs** for troubleshooting

## Quick start (download release package)

Don't want to build from source? Download a pre-packaged release, extract it and follow the steps below. The release package is self-contained — it includes the backend binary, frontend build artifacts, config files, database migration scripts and Docker deployment files. Just modify the config and start.

### 1. Download the release package

Go to the [Releases page](https://github.com/ilaziness/orange-tv/releases) and download the package for your platform:

- Linux: `orange-tv-<version>-linux-amd64.tar.gz`
- Windows: `orange-tv-<version>-windows-amd64.zip`

### 2. Extract

```sh
# Linux
tar -xzf orange-tv-<version>-linux-amd64.tar.gz
cd orange-tv-<version>/

# Windows
# Right-click the zip → Extract here, then enter the extracted directory
```

Directory structure after extraction:

```text
.
├── orange-tv                  # backend binary (Linux/macOS)
├── orange-tv.exe              # backend binary (Windows)
├── configs/                   # config files
│   ├── config.yaml            # default/sample config (with full field docs)
│   └── config.prod.yaml       # production config (overridable via env vars)
├── migrations/                # database migration scripts
├── web/
│   ├── client/                # client frontend build artifacts
│   └── admin/                 # admin frontend build artifacts
├── nginx/nginx.conf           # nginx sample config (client 80 / admin 81)
├── .env.example               # Docker Compose env var sample
├── Dockerfile                 # all-in-one image build file
├── docker-compose.yml         # one-click Docker Compose (app + MySQL)
└── docker-entrypoint.sh       # container entrypoint script
```

### 3. Run

#### Option 1: Docker Compose one-click deployment (recommended)

> Prerequisite: Docker and Docker Compose installed on the server (Docker 20.10+).

```sh
# Copy the env sample and modify as needed (DB credentials, JWT_SECRET, ports, etc.)
cp .env.example .env

# Start app + MySQL (first run auto-creates the database and runs migrations)
docker compose up -d

# Check logs to confirm it started cleanly
docker compose logs -f app
```

After startup:

- Client: `http://<host>:80`
- Admin: `http://<host>:81`

#### Option 2: Bare metal (Linux)

1. Edit `configs/config.prod.yaml` to set the database connection and other info (or override via env vars).
2. Run database migrations:

   ```sh
   ./orange-tv migrate up -c configs/config.prod.yaml
   ```

3. Start the backend:

   ```sh
   ./orange-tv serve -c configs/config.prod.yaml
   ```

4. Serve the frontend with nginx: refer to `nginx/nginx.conf`, adjust static resource paths, then `nginx -s reload`.

#### Option 3: Bare metal (Windows)

1. Edit `configs\config.prod.yaml` to set the database connection and other info.
2. Run database migrations:

   ```sh
   orange-tv.exe migrate up -c configs\config.prod.yaml
   ```

3. Start the backend:

   ```sh
   orange-tv.exe serve -c configs\config.prod.yaml
   ```

4. Serve `web\client` and `web\admin` static assets with IIS or another web server, reverse-proxying `/api` to `http://127.0.0.1:8080`.

### 4. Initialize the admin account

After the service starts and migrations complete, create a super admin account to log in to the admin backend. Replace `your_password` with your own password.

```sh
# Docker Compose
docker compose exec app /app/orange-tv admin create --username admin --password your_password --email admin@example.com

# Bare metal (Linux)
./orange-tv admin create --username admin --password your_password --email admin@example.com -c configs/config.prod.yaml

# Bare metal (Windows)
orange-tv.exe admin create --username admin --password your_password --email admin@example.com -c configs\config.prod.yaml
```

On success it prints something like `Created super_admin: id=1 username=admin`. Log in to the admin backend at `http://<host>:81` and start entering or collecting videos.

> The `README.md` bundled in the release package has more detailed deployment notes.

## How do I get started? (build from source)

If you are a developer or operator and want to build from source, deploy as follows:

1. Prepare a server with Go 1.26.4+ and MySQL 8.x/9.x installed
2. Clone this repo: `git clone https://github.com/ilaziness/orange-tv.git`
3. Follow the [deployment guide](docs/deployment.md) to configure, migrate and start
4. Create the first admin account via CLI, then log in to the admin backend to start entering or collecting videos

For detailed installation, configuration, commands and technical notes, see the [technical docs](docs/README.md).

## Documentation index

| Document | Content |
| ------ | ------ |
| [Technical docs](docs/README.md) | Project structure, configuration, commands, database, observability |
| [Deployment guide](docs/deployment.md) | Single-instance, multi-instance and Docker deployment |
| [Module usage](docs/module-usage.md) | How to select and remove unneeded service modules |
| [Observability guide](docs/observability.md) | Logging, tracing and metrics configuration |

## Disclaimer

This project is for learning and technical research purposes only. It does not provide, host or distribute any video content.

- The collection feature is for technical demonstration only. Users must ensure their collection activities comply with applicable laws and the source site's terms of use.
- Users are solely responsible for the copyright legality of collected content and the operation of their deployed sites.
- The project author is not liable for any legal disputes arising from the use or misuse of this system.

## License

This project is open-sourced under the [MIT License](LICENSE).
