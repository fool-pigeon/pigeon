the script is a [`expr`](https://expr-lang.org/) expression that returns bool

if your script not returns bool,not returns any value or get a syntax error,the test will be failed.

`expr` supports variable injection;
for the following request, we inject

## REST:
### pre-request
- [req](https://pkg.go.dev/resty.dev/v3#Request)
  
  It is worth noting
  that Go converts headers to standard case by default. If you wish to use a custom case (such as `"low_header"`), please use [`SetHeaderVerbatim`](https://pkg.go.dev/resty.dev/v3#Request.SetHeaderVerbatim)

  for example:
  ```
  req.SetHeaderVerbatim("lower_key", "value")
  ```

- [fake](https://github.com/jaswdr/faker)

  for example:
  ```
  fake.Person().Name()
  ```
### test
- [res](https://pkg.go.dev/resty.dev/v3#Response)


  for example has a JSON response like:
  ```
  {
    key:"value"
  }
  ```
  you can assert this:
  ```
  fromJSON(res.String()).key == "value"
  ```
- [fake](https://github.com/jaswdr/faker)

  for example:
  ```
  fake.Person().Name()
  ```
