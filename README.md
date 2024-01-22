# back-at - Show Progress Bar Until Given Time

[![Tests](https://github.com/tebeka/back-at/actions/workflows/test.yml/badge.svg)](https://github.com/tebeka/back-at/actions/workflows/test.yml)


## Usage

```
$ back-at -h
usage: back-at HH:MM (or HH:MMpm)
```

## Example

```
$ back-at 17:00
☕ ████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  16%
```

## Install

You can get the tool from the [GitHub release section](https://github.com/back-at/expmod/releases), or:

```
$ go install github.com/tebeka/back-at@latest
```

Make sure `$(go env GOPATH)/bin` is in your `$PATH`.
