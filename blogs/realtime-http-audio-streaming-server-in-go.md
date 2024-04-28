---

date: Jan-9-2024

---

# A Guide to Building a Realtime HTTP Audio Streaming Server in Go

Sometimes, I stumble upon internet radio stations that feature a diverse selection of music, often overlooked on mainstream streaming platforms. Finding unexpectedly great tracks fairly frequently, I've grown to appreciate this mode of audio consumption a lot more than I used to. 
Generally, while software like [Icecast](https://icecast.org/) implements it's own custom protocol for realtime audio streaming, streaming directly in through HTTP is relatively simpler and more accessible. This guide will delve into building such a server in pure go.

## Pre-requisites

- Go >= v1.21 (preferred)
- A web browser or a url music player that supports the AAC audio format.
- An audio file - encode it as an AAC audio file.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Managing Connections](#managing-connections)
3. [Stream Goroutine](#stream-goroutine)
4. [Starting the HTTP Server](#starting-the-http-server)

## Getting started

First, initialize the project and create main.go:

```.
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

In main.go, import all the required packages:

```go
import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)
```

Inside the main function, let's add some command line arguments to parse the stream file and open it with os.Open():

```go
fname := flag.String("filename", "file.aac", "path of the audio file")
flag.Parse()

file, err := os.Open(*fname)
if err != nil {
	log.Fatal(err)
}
```

Since the stream will have to be duplicated on each iteration, it is going to be stored as a slice of bytes.

```go
ctn, err := io.ReadAll(file)
if err != nil {
	log.Fatal(err)
}
```

**NOTE:**
1. For the sake of convenience, I will use a single audio file looping continuously for the livestream.
2. We are using AAC as our preferred file format due to it having self contained frames and not having any metadata.

## Managing Connections

It is outlandishly resource-intensive to create a new audio stream for each listener that connects to our server. Therefore, our server has to be designed to serve multiple clients with a single audio stream.

Since we have **multiple** clients, it is important to open, keep track of, and close of their connections to audio stream when necessary.

To achieve this, we'll create a connection pool:

```go
type ConnectionPool struct {
	bufferChannelMap map[chan []byte]struct{}
	mu               sync.Mutex
}
```

This ``ConnectionPool`` struct has two fields, a ``bufferChannelMap`` which is hashset that stores a channel to send and receive audio buffers, as well as ``sync.Mutex`` to guard the hashset so as to avoid race conditions during concurrent access.

We can implement the methods to add, delete and broadcast over these connections:

```go
func (cp *ConnectionPool) AddConnection(bufferChannel chan []byte) {

	defer cp.mu.Unlock()
	cp.mu.Lock()

	cp.bufferChannelMap[bufferChannel] = struct{}{}

}

func (cp *ConnectionPool) DeleteConnection(bufferChannel chan []byte) {

	defer cp.mu.Unlock()
	cp.mu.Lock()

	delete(cp.bufferChannelMap, bufferChannel)

}

func (cp *ConnectionPool) Broadcast(buffer []byte) {

	defer cp.mu.Unlock()
	cp.mu.Lock()

	for bufferChannel, _ := range cp.bufferChannelMap {
		clonedBuffer := make([]byte, 4096)
		copy(clonedBuffer, buffer)
		
		select {

		case bufferChannel <- clonedBuffer:

		default:

		}

	}

}
```

In the ``Broadcast()`` method, we iterate over each buffer channel and perform a non-blocking send. This will make sure that the stream is not bottlenecked by a slow write on an individual connection. We clone the buffer each time to avoid race conditions when it will eventually be read.

Along with the above methods, let's also create a function ``NewConnectionPool()`` to initialize a connection pool:

```go
func NewConnectionPool() *ConnectionPool {
	
	bufferChannelMap := make(map[chan []byte]struct{})
	return &ConnectionPool{bufferChannelMap: bufferChannelMap}

}
```

## Stream Goroutine

For actually broadcasting the audio, let's create a ``stream()`` function. This will contain the main stream loop and another overarching loop which duplicates and restarts the audio stream as soon as it ends.

```go
func stream(connectionPool *ConnectionPool, content []byte) {

	for {
		// duplicates the stream and creates a new ticker
		for {

			// consumes the stream and uses connectionPool.Broadcast 
			// to brodcast it on every tick

		}
	
	}
}
```

In the outer loop:

```go
for {
	tempfile := bytes.NewReader(content)
	// clear() is a builtin function introduced in go 1.21. 
	// Reinitialize the buffer if on a lower version.
	clear(buffer) 
	ticker := time.NewTicker(time.Millisecond * 250)

	for range ticker.C {
		// inner loop
	}
}
```

Here we convert the audio content that we had earlier stored as a slice of bytes into an ``io.Reader`` by enclosing it in a ``bytes.Reader``. We store this as the temporary stream which is created every time when the stream in the inner loop has finished being read.

We empty out our buffer so it can be reused and create a ticker for 250 milliseconds so that the stream has an output delay and doesn't send too much data in small interval of time.

EDIT: As mentioned in the comments, you can calculate and adjust the delay according to this formula: tick_duration = track_duration * buffer_size / aac_file_size

In the inner loop:

```go
for range ticker.C {
	_, err := tempfile.Read(buffer)
	
	if err == io.EOF {
		ticker.Stop()
		break
	}
	connectionPool.Broadcast(buffer)
}
```

Here, we read from the temporary buffer and broadcast to the connection pool until it reaches EOF - signalling the end of the audio stream, after which we stop the ticker and break out of the inner loop to restart the stream.

## Starting the HTTP Server

Back in the main function, let's initialize our connection pool and start the stream.

```go
connPool := NewConnectionPool()
go stream(connPool, ctn)
```

Now, using ``net/http`` 's default Handler,  we listen for incoming requests on ``/``.

```go
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// handle request
}
```

We add the required headers:

```go
w.Header().Add("Content-Type", "audio/aac")
w.Header().Add("Connection", "keep-alive")
```

We access the ``ResponseWriter`` 's flusher to flush writes on every received buffer:

```go
flusher, ok := w.(http.Flusher)
if !ok {

	log.Println("Could not create flusher")

}
```

Now, we create our client buffer channel, add it to the connection pool, and listen for broadcasts indefinitely.

```go
bufferChannel := make(chan []byte)
connPool.AddConnection(bufferChannel)
log.Printf("%s has connected\n", r.Host)

for {

	buf := <-bufferChannel
	if _, err := w.Write(buf); err != nil {

		connPool.DeleteConnection(bufferChannel)
		log.Printf("%s's connection has been closed\n", r.Host)
		return

	}
	flusher.Flush()

}
```

When ``w.Write`` returns an error, we can assume that the connection has broken off and delete it from the pool.

Finally, let's start the server on port 8080:

```go
log.Println("Listening on port 8080...")
log.Fatal(http.ListenAndServe(":8080", nil))
```

If we access localhost:8080 through a music player or a browser, we can listen to our audio being streamed. If we open it with multiple tabs or player instances, it runs synchronously on each device.

## Conclusion

We've successfully created our realtime audio streaming server. The code has been further optimized and is available on [Github](https://github.com/icelain/radio). A demo is available [here](https://ice.lqorg.com/music/stream)
