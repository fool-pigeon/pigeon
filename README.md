[![Go Reference](https://pkg.go.dev/badge/example.com.svg)](https://pkg.go.dev/example.com)

Project Structure

```
.
├── .conf
├── .env
├── f1
│   ├── r1.toml
│   └── r2.toml
├── f2
│   ├── .env
│   ├── f2_1
│   │   ├── .env
│   │   ├── r1.toml
│   │   └── r2.toml
│   └── {r1.toml}
├── .pg
├── r1.toml
└── r2.toml

```

cli functions
new - create a request
config - create a config for root project
env - set the env of collection
run - run the collections,requests