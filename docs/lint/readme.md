# How to Lint

We're using **golangci-lint** in our pipeline to integrate linting into our workflow. This will help us catch errors early and keep our codebase clean. You'll want to manually run **golangci-lint** and fix any linting errors before commiting to main.

````shell
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run
golangci-lint run --fix     # Try this first to autofix
gofmt -w .                  # Then fix remaining formatting
````

extra linters

```shell
go install mvdan.cc/gofumpt@latest
go install github.com/segmentio/golines@latest
golines --max-len=100 --base-formatter=gofumpt -w
```