# How to Lint

Linting is the process of running a program that will analyse code for potential errors. 
This is a good practice to ensure that the code is clean and free of errors. 
Linting can help catch errors early on in the development process, which can save time and effort in the long run.

````shell
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run
golangci-lint run --fix     # Try this first to autofix
gofmt -w .                  # Then fix remaining formatting
````