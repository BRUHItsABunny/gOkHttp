# Golang HTTP Client

[![Go Report Card](https://goreportcard.com/badge/BRUHItsABunny/gOkHttp)](https://goreportcard.com/report/BRUHItsABunny/gOkHttp)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/BRUHItsABunny/gOkHttp/master/LICENSE)

## Introduction

gOkHttp is HTTP Client written in Golang inspired by Java's (now Kotlin's) [OkHTTP](https://github.com/square/okhttp) and also slightly inspired by Python's [Requests](https://github.com/kennethreitz/requests). 

This library has been written with handling cookies (in-memory, on-disk, encrypted on-disk), headers, parameters, post fields, post multipart bodies, post files and SSL pinning in mind.

* Inspired by okhttp
    * SSL Pinning implementation and ease of using it
* Inspired by requests
    * Response processing (eg: `gokhttp.HttpResponse.Text()` to return response body as a string)
    * Making requests with headers and parameters as maps (there where in Python it's dicts)

##### Warning

This library is probably not production ready yet, most of the core was coded in under 2 hours at 1 AM. Proceed with caution, this library was made to fit MY needs and therefore may be structured in a weird way or incomplete (feel free to fix it in a pull request).

## Dependencies

Only external dependencies listed, much thanks to all of the repositories listed here!

* cookies/jar.go
    * `github.com/juju/go4/lock`
* response.go
    * `github.com/anaskhan96/soup` (for parsing HTML)
    * `github.com/beevik/etree` (for parsing XML)
    * `github.com/buger/jsonparser` (for parsing JSON)

Also thanks to `github.com/juju/persistent-cookiejar` as the coookie jar implementation is pretty much that code + AES encryption capabilities.

## Installation

The package can be installed using "go get".

```bash
go get -u github.com/BRUHItsABunny/gOkHttp
```

## Usage

There is a folder filled with examples [here](https://github.com/BRUHItsABunny/gOkHttp/tree/master/_examples/)

* [Simple request](https://github.com/BRUHItsABunny/gOkHttp/tree/master/_examples/simple_request.go)
* [Simple GET request](https://github.com/BRUHItsABunny/gOkHttp/tree/master/_examples/simple_get.go)
* [Simple POST request](https://github.com/BRUHItsABunny/gOkHttp/tree/master/_examples/simple_post.go)
* [Data POST request](https://github.com/BRUHItsABunny/gOkHttp/tree/master/_examples/data_post.go)
* [Multipart POST request](https://github.com/BRUHItsABunny/gOkHttp/tree/master/_examples/multipart_post.go)
* [Multipart file upload](https://github.com/BRUHItsABunny/gOkHttp/tree/master/_examples/multipart_file_post.go)
* [File upload](https://github.com/BRUHItsABunny/gOkHttp/tree/master/_examples/file_post.go)
* [SSL pinning](https://github.com/BRUHItsABunny/gOkHttp/tree/master/_examples/ssl_pinning.go)

## Contributions

Feel free to fork this repository and open up pull requests.

## Todo

* Built in HTTP3 support (as a switch)
* Built in HLS downloader
* Easier access to cookies
* Better Unmarshalling, support for custom unmarshallers
