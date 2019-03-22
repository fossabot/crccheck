# CRC Check

[![Build Status](https://travis-ci.com/dnaka91/crccheck.svg?branch=master)](https://travis-ci.com/dnaka91/crccheck)
[![codecov](https://codecov.io/gh/dnaka91/crccheck/branch/master/graph/badge.svg)](https://codecov.io/gh/dnaka91/crccheck)

A simple tool to verify CRC32 hashes in filenames against their content.

## Installation

#### Using Go Modules

The preferred way of installing `crccheck` is through Go Modules. Before proceeding please first install
[Mage](https://magefile.org). Then simply clone the project and run `mage install` inside.

```
git clone https://github.com/dnaka91/crccheck
cd crccheck
mage install
```

#### Using GOPATH

You can also install the application with the classical GOPATH approach. This doesn't use **Mage** or any dependency
management, but there is a chance for dependencies to have breaking changes and the build may fail.

```
go get -d github.com/dnaka91/crccheck
```

#### Using Docker

If you prefer to use Docker to run the application, there is also an image available.
Make sure to map the current working directory (or any folder where you want to search for files) to the
folder `/tmp`.

```
docker run --rm -it -v $(pwd):/tmp dnaka91/crccheck
```

## Usage Example

Let's assume we have a file with some content and the hash value set to zeros in the file name.
On first run the check will detect a **MISMATCH** as content and hash don't match.

After running again but with the argument `-u`, the hash will be updated.

The next time we run `crccheck` hash value and content match.

```
$ crccheck
file[00000000].txt - MISMATCH
$ crccheck -u
file[00000000].txt - UPDATED
$ crccheck
file[6515BCC3].txt - OK
```

### Help

To print further options and descriptions run crccheck with the argument `-h`.

```
$ crccheck -h
```

## License

This project is licensed under **Apache 2.0**.
