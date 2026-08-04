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

## How do I get started?

If you are a developer or operator, deploy as follows:

1. Prepare a server with Go 1.26.4+ and MySQL 8.x/9.x installed
2. Clone this repo: `git clone https://github.com/ilaziness/orange-tv.git`
3. Follow the [deployment guide](docs/deployment.md) to configure, migrate and start
4. Create the first admin account via CLI, then log in to the admin backend to start entering or collecting videos

For detailed installation, configuration, commands and technical notes, see the [technical docs](docs/README.md).

## Documentation index

| Document | Content |
| ------ | ------ |
| [Technical docs](docs/README.md) | Project structure, configuration, commands, database, observability |
| [Product requirements](docs/PRD.md) | Full product feature requirements and API design |
| [Deployment guide](docs/deployment.md) | Single-instance, multi-instance and Docker deployment |
| [Module usage](docs/module-usage.md) | How to select and remove unneeded service modules |
| [Observability guide](docs/observability.md) | Logging, tracing and metrics configuration |

## License

MIT License
