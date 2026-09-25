[![Go Reference](https://pkg.go.dev/badge/example.com.svg)](https://pkg.go.dev/example.com)

Example Project Structure

```
.
├── .pg
│   ├── envs
│   │   ├── dev
│   │   └── prod
│   ├── config.toml
│   ├── run-dev.toml
│   └── run.toml
├── dir
│   └── r1.toml
├── r1.toml
└── r2.toml
```

## Install

### CLI
```bash
go install github.com/fool-pigeon/pigeon/cmd/pg
```
### GUI
check the release

## TODO: Feature
* User Interface
  * [x] cli
  * [ ] gui
    * theme
      * [ ] a pretty default theme
      * [ ] theme for codemirror editors
      * [ ] dark mode and change it responsively
    * [ ] global file search filter
    * [ ] global key shortcuts bindings
    * [ ] request's send and test history
  * [ ] vscode-extension
* Protocols
  * [x] HTTP
  * [ ] SSE
  * [ ] gRPC
  * [ ] WS
  * [ ] GQL
  * [ ] ...
* Request Editor
  * [ ] option disable fields such as disable some headers
* Env Management
  * [ ] global environment(Specify the path. env)
  * [ ] collection environment
  * [ ] optional to use os' environment
  * [ ] environment on command's flags
* Test Feature
  * [x] Script inject
  * [ ] Script syntax highlighting 
  * [ ] Script autocompletion
  * [ ] Data Drive Test
  * [ ] Pre Request
  * [ ] Limit Concurrent
  * [ ] Set up synchronous running sequence
* Collection type conversion
  * [ ] OpenAPI
  * [ ] Postman
  * [ ] Apifox
  * [ ] Bruno
* LLM Feature(do not intergration in main version,because it's probably a redundant function.)
  * [ ] Optional LLM Integration to generate tests for only extra GUI version
  * [ ] auto create collection or request
  * [ ] auto run test


## TODO: For Project
* ci
  * [ ] run test and assert
  * [ ] auto build & release