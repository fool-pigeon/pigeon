<p align="center">
  <img width="200px" alt="pigeon-icon"
    src="./assets/pigeon-icon.webp" />
</p>
<h1 align="center">
  pigeon
</h1>
<p align="center">
    A simple API client – built with Wails3, Go, and Svelte5.
</p>
<br>

## Install

### CLI
```bash
go install github.com/fool-pigeon/pigeon/cmd/pigeon
```
### GUI
check in the release

## Features
`piegon` is a fully offline client that uses .toml files to configure project settings and individual requests.

You do not need to purchase the premium version of piegon to access any specific features, as piegon has no pricing tiers; all its current and future features are completely free.

### Performance
- Small software footprint, thanks to being built with [wails3](https://v3.wails.io/)
- Requests run concurrently by default

### Supported request types
-  REST
<!-- -  SSE -->
<!-- -  gRPC -->
<!-- -  WebSocket -->
<!-- -  GraphQL -->

### Request Task Execution Features
- Limit on concurrent requests
<!-- - Chained requests -->
<!-- - Configure synchronous request order -->

### Testing
- Script assertion testing with [`expr`](https://expr-lang.org/)
<!-- - Data-driven testing -->

### Organisation and Collaboration
- All file system paths use relative paths by default, facilitating migration and collaboration
- Switchable environment variables for project collections

### Extensions and Customisation
<!-- - Plugin system -->



## Example Project Structure

a project collection will be like
```
.
├── .pg
│   ├── run  # store the run batch tasks
│   │   ├── dev.toml
│   │   └── prod.toml
│   └── conf  # store the different configurations
│       ├── config.toml
│       ├── dev.toml
│       └── prod.toml
│
├── .conf # a .toml file that include nested config of request and environment variables
├── dir
│   ├── .conf
│   └── r1.toml
├── r1.toml
└── r2.toml
```
