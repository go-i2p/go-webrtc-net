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

// RTCConn implements net.Conn over WebRTC
type RTCConn struct {
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

// RTCPacketConn implements net.PacketConn over WebRTC
type RTCPacketConn struct {
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

// RTCListener implements net.Listener over WebRTC
type RTCListener struct {
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
