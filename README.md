# errgroupcheck

![CI](https://github.com/alexbagnolini/errgroupcheck/workflows/CI/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/alexbagnolini/errgroupcheck)](https://goreportcard.com/report/github.com/alexbagnolini/errgroupcheck)
[![Coverage](https://coveralls.io/repos/github/alexbagnolini/errgroupcheck/badge.svg?branch=master)](https://coveralls.io/github/alexbagnolini/errgroupcheck?branch=master)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

errgroupcheck is a linter that checks that you are calling `errgroup.Wait()` when using `errgroup.Group` from the `golang.org/x/sync/errgroup` package.

## Installation

```
go install github.com/alexbagnolini/errgroupcheck@latest
```

## Why

When using `errgroup.Group` from the `golang.org/x/sync/errgroup` package, it is important to call `errgroup.Wait()` to ensure that all goroutines have finished before the main goroutine exits.

## Examples

### Good usage

```go
func errgroupWithWait() {
	eg := errgroup.Group{}

	eg.Go(func() error {
		return nil
	})

	eg.Go(func() error {
		return nil
	})

	// call to .Wait()
	// (you probably want to check / return the error)
	_ = eg.Wait()
}
```

### Bad usage

```go
func errgroupMissingWait() {
	eg := errgroup.Group{} // golangci-lint will report "errgroup 'eg' does not have Wait called"

	eg.Go(func() error {
		return nil
	})

	eg.Go(func() error {
		return nil
	})
}
```