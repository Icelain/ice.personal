---

date: 2024/10/21

---

# Fast Unix Sockets with Tokio

In a world full of TCP wrapping and UDP-based protocols, Unix sockets are often overlooked due to the platform-agnostic and system-independent nature of the latter. However, when services run together on the same Unix box, Unix sockets are worth exploring, as their low-level and low-overhead nature provides better performance than pure TCP streams and UDP datagrams.

This blog post provides a gentle introduction to Unix sockets through **Rust** using the **Tokio** async runtime.

## Table of Contents

1. [Setup and Dependencies](#setup-and-dependencies)
2. [UnixListener Server](#unixlistener-server)
3. [UnixStream Client](#unixstream-client)
4. [Wrapping it up with Boilerplate](#wrapping-it-up-with-boilerplate)

## Setup and Dependencies

We'll begin by setting up the project as any other Rust project:

```
cargo new your_project_name
```

Inside the Cargo.toml file, we'll add Tokio with these feature flags:
```toml
[dependencies]
tokio = { version = "1.40.0", features = ["rt-multi-thread", "fs", "net", "io-std", "io-util", "sync", "macros", "signal"] }
```

NOTE: It is more pragmatic to simply add:
```toml
features=["full"]
```
since we are using most of Tokio's features.

## UnixListener Server

First, let's include all the required libraries:
```rust
use std::env::args;
use std::path::Path;
use std::process::exit;
use tokio::io::{self, AsyncBufReadExt, AsyncReadExt, AsyncWriteExt, BufReader};
use tokio::net::{UnixListener, UnixStream};
use tokio::signal;
use tokio::sync::mpsc::{channel, Receiver};
```

Inside src/main.rs of the project, let's begin with the server.

```rust
async fn server(socket_path: String, mut shutdown_receiver: Receiver<()>) {}
```

Our server function signature is straightforward: we have `socket_path`, which is the path of the socket we'll be listening on, and a `shutdown_receiver` channel for listening to ctrl-c and gracefully shutting down our server.

Now we have:
```rust
let socket_path_buf = Path::new(&socket_path).to_path_buf();
let socket_path_buf_clone = socket_path_buf.clone(); // We'll be using the socket path buf again in a different Tokio task
let listener = UnixListener::bind(socket_path_buf).expect("Could not create unix socket");
```

We create a `socket_path_buf` of type `PathBuf` for our socket path and pass it to `UnixListener::bind`. This returns a result likely withholding our UnixSocket listener (possible errors can include the socket file already existing).

Let's spawn a new Tokio task to handle graceful shutdowns on `<ctrl-c>`:
```rust
tokio::spawn(async move {
    match shutdown_receiver.recv().await {
        Some(()) => {
            tokio::fs::remove_file(socket_path_buf_clone)
                .await
                .expect("Failed to remove socket file");

            exit(1);
        }
        None => {
            eprintln!(
                "received nothing from the shutdown receiver. This should not be possible"
            )
        }
    }
});
```

This task waits for the shutdown_receiver channel to return so it can clean up by removing the socket file and quit the program.
Now, for the main part:

```rust
while let Ok((mut stream, _)) = listener.accept().await {

    println!("Listening on {socket_path}");
    let mut buffer: [u8; 1024] = [0u8; 1024];
    tokio::spawn(async move {
        loop {
            match stream.read(&mut buffer).await {
                Ok(n) => {
                    if n == 0 {
                        break;
                    }

                    println!("client: {:?}", String::from_utf8_lossy(&buffer[..n]));
                }

                Err(e) => {
                    eprintln!("Error writing to client; error: {}", e);
                    break;
                }
            }
        }
    });
}
```

We continuously accept new streams with `while let Ok((mut stream, _)) = listener.accept().await{}`, allocating a 1KB buffer for each of them to read bytes to. Consequently, we spawn a task that reads bytes from the stream to the allocated buffer. If the number of bytes read is 0, it indicates that the client has disconnected, so we break from the loop and end the task. We do the same if there is an error while reading the stream.

With that, our Unix socket server is complete. Let's move on to implementing the client.

## UnixStream Client

The function signature for the client is similar:
```rust
async fn client(socket_path: String, mut shutdown_receiver: Receiver<()>) {}
```

We take in a socket_path to connect to and a shutdown_receiver channel to close the client orderly.
Inside our client function, we connect to the Unix socket server through the `UnixStream` struct and obtain a mutable reference to it so we can write to it later:

```rust
let mut unixstream = UnixStream::connect(Path::new(&socket_path)).await.expect("Could not connect to the socket path. Ensure that the path is correct and is being listened on.");
println!("Connected to {socket_path}");
```

Similar to our server, we spawn a task to gracefully shut down our client as well:
```rust
tokio::spawn(async move {
        match shutdown_receiver.recv().await {
            Some(()) => {
                println!("Shutting down the client");
                exit(1);
            }
            None => {
                eprintln!(
                    "received nothing from the shutdown receiver. This should not be possible"
                )
            }
        }
});
```

Let's initialize a handle to stdout and stdin:
```rust
let mut stdout = io::stdout();
let mut stdin_lines = BufReader::new(io::stdin()).lines();
```

Finally, let's open a loop where we read lines of text from stdin and write them to the socket connection:
```rust
loop {
    stdout.write(b"Text: ").await.unwrap();
    stdout.flush().await.unwrap();

    if let Some(line) = stdin_lines.next_line().await.unwrap() {
        unixstream.write(line.as_bytes()).await.unwrap();
    }
}
```

This completes the client implementation.

## Wrapping it up with Boilerplate

Now that we're left with our main function, we need to spawn our task which listens to and relays ctrl-c events and take in command-line arguments so we can run a client or a server depending on them:

```rust
#[tokio::main]
async fn main() {
    let mode = args().nth(1).unwrap();
    let socket_path = args().nth(2).unwrap();
    let (shutdown_sender, shutdown_receiver) = channel(1);

    tokio::spawn(async move {
        match signal::ctrl_c().await {
            Ok(()) => {
                shutdown_sender.send(()).await.unwrap();
            }
            Err(e) => {
                eprintln!("{}", e)
            }
        }
    });

    if mode.as_str() == "server" {
        server(socket_path, shutdown_receiver).await;
    } else if mode.as_str() == "client" {
        client(socket_path, shutdown_receiver).await;
    } else {
        println!("Provide valid operation");
    }
}
```

Let's try running our server:
```
cargo run --release server myunixsocket.sock
   Compiling unixtokio v0.1.0 (/Users/icell/Desktop/code/rust/unixtokio)
    Finished `release` profile [optimized] target(s) in 0.39s
     Running `target/release/unixtokio server myunixsocket.sock`
Listening on myunixsocket.sock
```

And connect to it using the client:
```
cargo run --release client myunixsocket.sock
    Finished `release` profile [optimized] target(s) in 0.01s
     Running `target/release/unixtokio client myunixsocket.sock`
Connected to myunixsocket.sock
Text: random text
Text: going through our unix socket
```

We'll see that our stream of bytes has reached the server:
```
Listening on myunixsocket.sock
client: "random text"
client: "going through our unix socket"
```

This completes our implementation of Unix domain sockets in Tokio. As always, the source code is available on [GitHub](https://github.com/Icelain/unixtokio).
