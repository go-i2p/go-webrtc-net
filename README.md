# go-webrtc-net

Go library providing standard Go network interfaces (net.Conn, net.PacketConn, net.Listener) over WebRTC connections.

## Installation

```bash
go get github.com/go-i2p/go-webrtc-net
```

## Features

- Standard Go network interfaces over WebRTC
- Support for both reliable (TCP-like) and unreliable (UDP-like) modes
- Thread-safe implementation
- Context-aware connection management
- Built on pure Go WebRTC implementation ([pion/webrtc](https://github.com/pion/webrtc))

## Usage

```go
// Create a WebRTC connection
conn, err := webrtc.DialConn(underlying, "remote-peer-address")
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Use it like a regular net.Conn
_, err = conn.Write([]byte("Hello WebRTC!"))
```

```go
// Create a WebRTC listener
listener, err := webrtc.Listen(tcpListener)
if err != nil {
    log.Fatal(err)
}
defer listener.Close()

// Accept connections
conn, err := listener.Accept()
```

```go
// Create a WebRTC packet connection
pconn, err := webrtc.DialPacketConn(underlying, "remote-peer-address")
if err != nil {
    log.Fatal(err)
}
defer pconn.Close()
```

## Testing

```bash
go test ./...
```

## License

MIT License

## Contributing

1. Fork the repository
2. Create your feature branch
3. Submit a pull request