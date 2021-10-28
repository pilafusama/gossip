package transport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/pilafusama/gossip/base"
	"github.com/pilafusama/gossip/log"
)

var (
	wsSubProtocol = "sip"
)

type WsConn struct {
	net.Conn
	client bool
}

func (wc *WsConn) Read(b []byte) (n int, err error) {
	var msg []byte
	var op ws.OpCode
	if wc.client {
		msg, op, err = wsutil.ReadServerData(wc.Conn)
	} else {
		msg, op, err = wsutil.ReadClientData(wc.Conn)
	}
	if err != nil {
		// handle error
		var wsErr wsutil.ClosedError
		if errors.As(err, &wsErr) {
			return n, io.EOF
		}
		return n, err
	}
	if op == ws.OpClose {
		return n, io.EOF
	}
	copy(b, msg)
	return len(msg), err
}

func (wc *WsConn) Write(b []byte) (n int, err error) {
	if wc.client {
		err = wsutil.WriteClientMessage(wc.Conn, ws.OpText, b)
	} else {
		err = wsutil.WriteServerMessage(wc.Conn, ws.OpText, b)
	}
	if err != nil {
		// handle error
		var wsErr wsutil.ClosedError
		if errors.As(err, &wsErr) {
			return n, io.EOF
		}
		return n, err
	}
	return len(b), nil
}

type Ws struct {
	connTable
	listeningPoints []*net.TCPListener
	network         string
	output          chan base.SipMessage
	stop            bool
	dialer          ws.Dialer
	up              ws.Upgrader
}

func NewWs(output chan base.SipMessage) (*Ws, error) {
	w := Ws{
		network: "ws",
		output:  output,
	}
	w.listeningPoints = make([]*net.TCPListener, 0)
	w.connTable.Init()
	w.dialer.Protocols = []string{wsSubProtocol}
	w.dialer.Timeout = time.Minute
	w.up.Protocol = func(val []byte) bool {
		return string(val) == wsSubProtocol
	}
	return &w, nil
}

func (w *Ws) Listen(address string) error {
	var err error = nil
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return err
	}

	lp, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	w.listeningPoints = append(w.listeningPoints, lp)
	go w.serve(lp)

	// At this point, err should be nil but let's be defensive.
	return err
}

func (w *Ws) Send(addr string, msg base.SipMessage) error {
	conn, err := w.getConnection(addr)
	if err != nil {
		return err
	}
	err = conn.Send(msg)
	return err
}

func (w *Ws) getConnection(addr string) (*connection, error) {
	conn := w.connTable.GetConn(addr)
	if conn == nil {
		log.Debug("No stored connection for address %s; generate a new one", addr)
		raddr, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			return nil, err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		url := fmt.Sprintf("%s://%s", w.network, raddr)
		baseConn, _, _, err := w.dialer.Dial(ctx, url)
		if err == nil {
			baseConn = &WsConn{
				Conn:   baseConn,
				client: true,
			}
		} else {
			if baseConn == nil {
				return nil, fmt.Errorf("dial to %s %s: %w", w.Network(), raddr, err)
			}
			log.Warn("fallback to TCP connection due to WS upgrade error: ", err.Error())
		}

		conn = NewConn(baseConn, w.output)
	}

	w.connTable.Notify(addr, conn)
	return conn, nil
}

func (w *Ws) IsStreamed() bool {
	return true
}

func (w *Ws) serve(listeningPoint *net.TCPListener) {
	log.Info("Begin serving TCP on address " + listeningPoint.Addr().String())

	for {
		baseConn, err := listeningPoint.Accept()
		if err != nil {
			log.Severe("Failed to accept TCP conn on address " + listeningPoint.Addr().String() + "; " + err.Error())
			continue
		}

		_, err = w.up.Upgrade(baseConn)
		if err != nil {
			log.Warn("fallback to simple TCP connection due to WS upgrade error: ", err.Error())
		} else {
			baseConn = &WsConn{
				Conn:   baseConn,
				client: false,
			}
		}
		conn := NewConn(baseConn, w.output)
		log.Debug("Accepted new TCP conn %p from %s on address %s", &conn, conn.baseConn.RemoteAddr(), conn.baseConn.LocalAddr())
		w.connTable.Notify(baseConn.RemoteAddr().String(), conn)
	}
}

func (w *Ws) Stop() {
	w.connTable.Stop()
	w.stop = true
	for _, lp := range w.listeningPoints {
		lp.Close()
	}
}

func (w *Ws) Network() string {
	return strings.ToUpper(w.network)
}
