[![Go Reference](https://pkg.go.dev/badge/example.com.svg)](https://pkg.go.dev/example.com)

Example Project Structure

```
.
├── .pg
│   ├── config-dev.toml
│   ├── config.toml
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