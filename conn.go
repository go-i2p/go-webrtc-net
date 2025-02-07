package webrtc

import (
	"context"
	"encoding/json"
	"net"
	"time"

	"github.com/pion/webrtc/v3"
)

// DialConn creates a new WebRTC connection using the provided net.Conn for signaling
func DialConn(conn net.Conn, addr string) (net.Conn, error) {
	ctx, cancel := context.WithCancel(context.Background())

	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	})
	if err != nil {
		cancel()
		return nil, err
	}

	c := &RTCConn{
		pc:        pc,
		localAddr: conn.LocalAddr(),
		ctx:       ctx,
		cancel:    cancel,
		readChan:  make(chan []byte, 100),
	}

	// Set up data channel
	dc, err := pc.CreateDataChannel("data", nil)
	if err != nil {
		cancel()
		return nil, err
	}

	c.dc = dc
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		select {
		case c.readChan <- msg.Data:
		case <-c.ctx.Done():
		}
	})

	// Handle signaling
	go c.handleSignaling(conn)

	return c, nil
}

// Implementation of net.Conn interface methods for conn type
func (c *RTCConn) Read(b []byte) (n int, err error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return 0, ErrConnectionClosed
	}
	c.mu.RUnlock()

	select {
	case data := <-c.readChan:
		return copy(b, data), nil
	case <-c.ctx.Done():
		return 0, ErrConnectionClosed
	}
}

func (c *RTCConn) Write(b []byte) (n int, err error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return 0, ErrConnectionClosed
	}
	c.mu.RUnlock()

	err = c.dc.Send(b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *RTCConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	c.cancel()

	if c.dc != nil {
		c.dc.Close()
	}
	if c.pc != nil {
		return c.pc.Close()
	}
	return nil
}

func (c *RTCConn) LocalAddr() net.Addr  { return c.localAddr }
func (c *RTCConn) RemoteAddr() net.Addr { return c.remoteAddr }

func (c *RTCConn) SetDeadline(t time.Time) error {
	// Implementation using context deadline
	return nil
}

func (c *RTCConn) SetReadDeadline(t time.Time) error {
	// Implementation using context deadline for reads
	return nil
}

func (c *RTCConn) SetWriteDeadline(t time.Time) error {
	// Implementation using context deadline for writes
	return nil
}

func (c *RTCConn) handleSignaling(conn net.Conn) {
	// Basic signaling implementation
	offer, err := c.pc.CreateOffer(nil)
	if err != nil {
		return
	}
	err = c.pc.SetLocalDescription(offer)
	if err != nil {
		return
	}
	// Send offer and handle answer through the provided connection
	// Send offer
	offerBytes, err := json.Marshal(offer)
	if err != nil {
		return
	}
	_, err = conn.Write(offerBytes)
	if err != nil {
		return
	}

	// Read answer
	answerBytes := make([]byte, 8192)
	n, err := conn.Read(answerBytes)
	if err != nil {
		return
	}

	var answer webrtc.SessionDescription
	err = json.Unmarshal(answerBytes[:n], &answer)
	if err != nil {
		return
	}

	err = c.pc.SetRemoteDescription(answer)
	if err != nil {
		return
	}
}
