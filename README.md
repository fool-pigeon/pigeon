[![Go Reference](https://pkg.go.dev/badge/example.com.svg)](https://pkg.go.dev/example.com)

Example Project Structure

```
.
├── .pg
│   ├── config-dev.toml
│   ├── config.toml
│   ├── run-dev.toml
│   ├── run.toml
│   ├── .env
│   └── .env-dev
├── .env
├── dir1
│   ├── .env
│   └── r1.toml
├── dir2
│   ├── dir2.1
│   └── dir2.2
│       ├── r1.toml
│       └── r2.toml
├── r1.toml
└── r2.toml
```

pg-cli commands
* init
  * --tags, Set additional configuration with name suffix tags
* env
  * init
    * --tags, Set additional configuration with name suffix tags
  * ls
    * args: directory(default current directory)
    * --tag, Print env with name suffix tag
* new
  * args: rName...
  * --protocol,-p
* run
  * args: paths...
  * --tag
* info
  * args: fName

## TODO: Feature
* User Interface
  * [x] cli
  * [ ] gui
    * [ ] global file search filter
    * [ ] global key shortcuts bindings
    * [ ] theme and change it responsively
    * [ ] request's send and test history
  * [ ] vscode-extension
* Protocols
  * [x] HTTP
  * [ ] SSE
  * [ ] gRPC
  * [ ] WS
  * [ ] GQL
  * [ ] ...
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