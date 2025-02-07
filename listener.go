package webrtc

import (
	"context"
	"net"

	"github.com/pion/webrtc/v3"
)

// Listen creates a new WebRTC listener using the provided net.Listener for signaling
func Listen(lstn net.Listener) (net.Listener, error) {
	ctx, cancel := context.WithCancel(context.Background())

	l := &listener{
		underlying: lstn,
		config: &webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{URLs: []string{"stun:stun.l.google.com:19302"}},
			},
		},
		acceptChan: make(chan net.Conn),
		ctx:        ctx,
		cancel:     cancel,
	}

	go l.acceptLoop()
	return l, nil
}

func (l *listener) Accept() (net.Conn, error) {
	select {
	case conn := <-l.acceptChan:
		return conn, nil
	case <-l.ctx.Done():
		return nil, ErrConnectionClosed
	}
}

func (l *listener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil
	}

	l.closed = true
	l.cancel()
	return l.underlying.Close()
}

func (l *listener) Addr() net.Addr {
	return l.underlying.Addr()
}

func (l *listener) acceptLoop() {
	for {
		conn, err := l.underlying.Accept()
		if err != nil {
			l.Close()
			return
		}
		select {
		case l.acceptChan <- conn:
		case <-l.ctx.Done():
			return
		}
	}
}
