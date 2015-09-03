# pkglines

[![Build Status](https://travis-ci.org/DeedleFake/pkglines.svg)](https://travis-ci.org/DeedleFake/pkglines)

pkglines is a simple app designed to count the lines of code in a Go
package and all of its dependencies. It is currently in early
development.

## Installation

```bash
> go get -u -v github.com/DeedleFake/pkglines
```

## Usage

Prints the number of lines of code in pkglines. pkglines has no
non-standard library dependencies, so this only prints pkglines's size.

```bash
> pkglines github.com/DeedleFake/pkglines
```

Prints the number of lines of code in pkglines and its dependencies.

```bash
> pkglines -std github.com/DeedleFake/pkglines
```
