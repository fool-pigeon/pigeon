Edit a [dotenv file](https://www.dotenv.org/docs/) in .pg/env and switch to it.

The contents of the file are parsed into a map of string key-value pairs, which are substituted when the file is loaded. We use `${}` as the environment variable template; templates cannot be nested.
For example,
the `env` content is
```
hello=world
```
When `${hello}` appears in the file content, it will be replaced with `world`.
For example, `${hello}${${hello}}` becomes `world${${hello}}`.