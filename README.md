# Golang HTTP Client

[![Go Report Card](https://goreportcard.com/badge/BRUHItsABunny/gOkHttp)](https://goreportcard.com/report/BRUHItsABunny/gOkHttp)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/BRUHItsABunny/gOkHttp/master/LICENSE)

## Introduction

gOkHttp is HTTP Client written in Golang inspired by Java's (now Kotlin's) [OkHTTP](https://github.com/square/okhttp) and also slightly inspired by Python's [Requests](https://github.com/kennethreitz/requests). 

This library has been written with handling cookies (in-memory, on-disk, encrypted on-disk), headers, parameters, post fields, post multipart bodies, post files and SSL pinning in mind.

* Inspired by okhttp
    * SSL Pinning implementation and ease of using it
* Inspired by requests
    * Response processing (eg: `responses.ResponseText(httpResp)` to return response body as a string)
    * Making requests with headers and parameters as maps (there where in Python it's dicts)

#### Features
This library has the following features in its ecosystem:
* Easy client construction
* Easy request construction
* Easy response parsing
* HTML parsing (and soon other advanced formats) with [goquery](https://github.com/PuerkitoBio/goquery) found [here](https://github.com/BRUHItsABunny/gOkHttp-advanced-responses)
* Cookie handling (in-memory, on-disk, encrypted on-disk) found [here](https://github.com/BRUHItsABunny/gOkHttp-cookiejar)
* SSL Pinning found [here](https://github.com/BRUHItsABunny/gOkHttp-sslpinning)
* HTTP3 client support (theoretical) found [here](https://github.com/BRUHItsABunny/gOkHttp-http3)
* Easy to use downloader implementations (multi-threaded, HLS stream downloader) found [here](https://github.com/BRUHItsABunny/gOkHttp-download)
* JA3 fingerprint spoofing found [here](https://github.com/BRUHItsABunny/gOkHttp-ja3spoof)

#### Why are some features split into another repository?
In order to keep the main library lightweight I needed these dependencies to be optional, if you need them you can import them as-needed.
This will result in smaller binary sizes and fewer dependencies to scan for licensing issues.

##### Warning

This library is a lot more stable than it was back when I first wrote it, though it's far from perfect so do not use it for mission critical stuff

## Installation

The package can be installed using "go get".

```bash
go get -u github.com/BRUHItsABunny/gOkHttp
```

## Usage

There is a file filled with examples [here](https://github.com/BRUHItsABunny/gOkHttp/blob/master/lib_test.go), and you can find the Go docs [here](https://pkg.go.dev/github.com/BRUHItsABunny/gOkHttp)


### NOTE for HLS streams
If the .ts file is broken, run it through FFmpeg in order to fix it real quick: `ffmpeg -i file.ts -c copy file_fixed.ts`

## Contributions

Feel free to fork this repository and open up pull requests.

## Todo

