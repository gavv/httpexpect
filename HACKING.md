# Hacking guidelines

## Developer dependencies

To develop httpexpect, you need:

* [golangci-lint](https://golangci-lint.run/usage/install/#local-installation)

* [stringer](https://github.com/golang/tools)

    `go install golang.org/x/tools/cmd/stringer@latest`

## Makefile targets

Re-generate, build, lint, and test everything:

```
make
```

Run tests:

```
make test
```

Run only short tests:

```
make short
```

Run gofmt:

```
make fmt
```

Run go generate:

```
make gen
```

Run go mod tidy:

```
make tidy
```

## Comment formatting

Exported functions should have documentation comments, formatted as follows:

* short function description, indented with one SPACE
* empty line
* optional details, indented with one SPACE
* empty line
* `Example:` line, indented with one SPACE
* empty line
* example code, indented with one TAB
* no more empty lines

**GOOD:**

```go
// Short function description.
//
// Optional details, probably multiple
// lines or paragraphs.
//
// Example:
//
//	exampleCode()
func MyFunction() { ... }
```

**BAD:** no space after `//`:

```go
//Short function description.
//
// Example:
//
//	exampleCode()
func MyFunction() { ... }
```

**BAD:** missing empty line before `Example:`

```go
// Short function description.
// Example:
//
//	exampleCode()
func MyFunction() { ... }
```

**BAD:** missing empty line after `Example:`

```go
// Short function description.
//
// Example:
//	exampleCode()
func MyFunction() { ... }
```

**BAD:** forgetting to indent example code:

```go
// Short function description.
//
// Example:
//
// exampleCode()
func MyFunction() { ... }
```

**BAD:** using spaces instead of TAB to indent example code:

```go
// Short function description.
//
// Example:
//
//  exampleCode()
func MyFunction() { ... }
```

**BAD:** extra empty line between comment and function:

```go
// Short function description.
//
// Example:
//
//	exampleCode()

func MyFunction() { ... }
```
