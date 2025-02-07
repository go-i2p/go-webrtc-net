package webrtc

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/pion/webrtc/v3"
)

var (
	ErrConnectionClosed = errors.New("connection closed")
	ErrInvalidAddress   = errors.New("invalid address")
)

// conn implements net.Conn over WebRTC
type conn struct {
	dc         *webrtc.DataChannel
	pc         *webrtc.PeerConnection
	localAddr  net.Addr
	remoteAddr net.Addr
	mu         sync.RWMutex
	closed     bool
	readChan   chan []byte
	ctx        context.Context
	cancel     context.CancelFunc
}

// packetConn implements net.PacketConn over WebRTC
type packetConn struct {
	pc            *webrtc.PeerConnection
	dc            *webrtc.DataChannel
	localAddr     net.Addr
	mu            sync.RWMutex
	closed        bool
	readChan      chan packet
	ctx           context.Context
	cancel        context.CancelFunc
	readDeadline  chan time.Time
	writeDeadline chan time.Time
}

// listener implements net.Listener over WebRTC
type listener struct {
	underlying net.Listener
	config     *webrtc.Configuration
	mu         sync.RWMutex
	closed     bool
	acceptChan chan net.Conn
	ctx        context.Context
	cancel     context.CancelFunc
}

// packet represents a UDP-like packet
type packet struct {
	data []byte
	addr net.Addr
}

var STUN_SERVER_URLS = []string{"stun:stun.l.google.com:19302"}
