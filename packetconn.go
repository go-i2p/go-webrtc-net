package webrtc

import (
	"context"
	"net"
	"time"

	"github.com/pion/webrtc/v3"
)

// Close implements net.PacketConn.
func (p *packetConn) Close() error {
	p.cancel()
	if p.dc != nil {
		if err := p.dc.Close(); err != nil {
			return err
		}
	}
	if p.pc != nil {
		return p.pc.Close()
	}
	return nil
}

// LocalAddr implements net.PacketConn.
func (p *packetConn) LocalAddr() net.Addr {
	return p.localAddr
}

// ReadFrom implements net.PacketConn.
func (p *packetConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	select {
	case <-p.ctx.Done():
		return 0, nil, p.ctx.Err()
	case pkt := <-p.readChan:
		n = copy(b, pkt.data)
		return n, pkt.addr, nil
	}
}

// SetDeadline implements net.PacketConn.
func (p *packetConn) SetDeadline(t time.Time) error {
	err1 := p.SetReadDeadline(t)
	err2 := p.SetWriteDeadline(t)
	if err1 != nil {
		return err1
	}
	return err2
}

// SetReadDeadline implements net.PacketConn.
func (p *packetConn) SetReadDeadline(t time.Time) error {
	if p.readDeadline == nil {
		p.readDeadline = make(chan time.Time, 1)
	}
	if !t.IsZero() {
		go func() {
			time.Sleep(time.Until(t))
			p.readDeadline <- t
		}()
	}
	return nil
}

// SetWriteDeadline implements net.PacketConn.
func (p *packetConn) SetWriteDeadline(t time.Time) error {
	if p.writeDeadline == nil {
		p.writeDeadline = make(chan time.Time, 1)
	}
	if !t.IsZero() {
		go func() {
			time.Sleep(time.Until(t))
			p.writeDeadline <- t
		}()
	}
	return nil
}

// WriteTo implements net.PacketConn.
func (p *packetConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	if p.dc == nil || p.dc.ReadyState() != webrtc.DataChannelStateOpen {
		return 0, net.ErrClosed
	}

	select {
	case <-p.ctx.Done():
		return 0, p.ctx.Err()
	case <-p.writeDeadline:
		return 0, context.DeadlineExceeded
	default:
		err = p.dc.Send(b)
		if err != nil {
			return 0, err
		}
		return len(b), nil
	}
}

// DialPacketConn creates a new WebRTC packet connection
func DialPacketConn(pconn net.PacketConn, raddr string) (net.PacketConn, error) {
	ctx, cancel := context.WithCancel(context.Background())

	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: STUN_SERVER_URLS},
		},
	})
	if err != nil {
		cancel()
		return nil, err
	}

	p := &packetConn{
		pc:        pc,
		localAddr: pconn.LocalAddr(),
		ctx:       ctx,
		cancel:    cancel,
		readChan:  make(chan packet, 100),
	}

	// Set up data channel with ordered: false for UDP-like behavior
	dc, err := pc.CreateDataChannel("data", &webrtc.DataChannelInit{
		Ordered: new(bool), // false for UDP-like behavior
	})
	if err != nil {
		cancel()
		return nil, err
	}

	p.dc = dc

	return p, nil
}
