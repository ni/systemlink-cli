# SystemLink Command-Line Interface

[![Build Status](https://github.com/ni/systemlink-cli/workflows/build/badge.svg)](https://github.com/ni/systemlink-cli/actions)
[![License](https://img.shields.io/github/license/ni/systemlink-cli)](https://github.com/ni/systemlink-cli/blob/master/LICENSE)

The systemlink-cli project is a simple command line interface over SystemLink services. It is implemented in golang and works natively on Windows, Linux and MacOS.

# How to use?

Take a look at the "[Getting Started](GettingStarted.md)" guide to learn about installing and using the SystemLink CLI.

# How to compile?

## Prerequisites

- Install golang compiler (https://golang.org/dl/)

```bash
sudo apt-get install golang-go
```

## Set up workspace and compile

The "build.sh" script downloads dependencies and builds the Linux, Windows and MacOS executables.

```bash
bin/build.sh
```

## How to run the tests?

```bash
bin/test.sh
```

## How to get code coverage results?

The following script calculates the code coverage results:

```bash
bin/coverage.sh
```

And this is how you can visualize the coverage results:

```bash
go tool cover -html=coverage.out
```

## How to (cross)-compile a single executable?

You can compile for the target setting the GOOS and GOARCH env variables. Here is a simple command to compile the executables for Linux on x86:

```bash
GOOS=linux GOARCH=386 go build -o build/systemlink cmd/main.go
build/systemlink tags get-tags
```