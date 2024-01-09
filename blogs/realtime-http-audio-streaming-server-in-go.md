# A Guide to Building a Realtime HTTP Audio Streaming Server in Go

I occasionally discover internet radio stations that offer a great assortment of music which usually goes unnoticed on traditional streaming platforms. I've grown to appreciate this mode of audio consumption a lot more than I used to. 
Generally, while software like [Icecast](https://icecast.org/) implements it's own custom protocol for realtime audio streaming, streaming directly in through HTTP also works really well.

## Pre-requisites
- Go >= v1.21 (preferred)
- A web browser or a url music player that supports the AAC audio format.
- An audio file - encode it as an AAC audio file.
## Table of Contents
[Getting Started](#getting-started)
## Getting started
First, initialize the project and create main.go:
```
go mod init radio
touch main.go	
```


Then, move your audio file to project directory. The directory structure should look like this:
```.
├── file.aac
├── go.mod
├── go.sum
└── main.go
```


