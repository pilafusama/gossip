package transport

import (
	"crypto/tls"
	"fmt"
	"github.com/pilafusama/gossip/log"
	"net"
	"time"

	"github.com/pilafusama/gossip/base"
)

type Wss struct {
	Ws
	listeningHandlers []*listenerHandler
	certPath, keyPath string
}

func NewWss(output chan base.SipMessage, certPath, keyPath string) (*Wss, error) {
	w := Wss{}
	w.certPath = certPath
	w.keyPath = keyPath
	w.network = "wss"
	w.output = output
	w.listeningHandlers = make([]*listenerHandler, 0)
	w.connTable.Init()
	w.dialer.Protocols = []string{wsSubProtocol}
	w.dialer.Timeout = time.Minute
	w.dialer.TLSConfig = &tls.Config{
		InsecureSkipVerify: false,
		//VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		//	return nil
		//},
	}
	w.up.Protocol = func(val []byte) bool {
		return string(val) == wsSubProtocol
	}
	return &w, nil
}

func (w *Wss) Listen(address string) error {
	var err error = nil
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return err
	}

	cert, err := tls.LoadX509KeyPair(w.certPath, w.keyPath)
	if err != nil {
		return fmt.Errorf("load TLS certficate %s: %w", w.certPath, err)
	}

	lp, err := tls.Listen("tcp", addr.String(), &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	if err != nil {
		return err
	}

	handler := &listenerHandler{listener: lp}
	w.listeningHandlers = append(w.listeningHandlers, handler)
	go w.serve(handler)

	// At this point, err should be nil but let's be defensive.
	return err
}

func (w *Wss) serve(handler *listenerHandler) {
	log.Info("Begin serving TCP on address " + handler.listener.Addr().String())

	for {
		baseConn, err := handler.listener.Accept()
		if err != nil {
			log.Severe("Failed to accept TCP conn on address " + handler.listener.Addr().String() + "; " + err.Error())
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

type listenerHandler struct {
	listener net.Listener
}

func (w *Wss) Stop() {
	w.connTable.Stop()
	w.stop = true
	for _, lp := range w.listeningHandlers {
		lp.listener.Close()
	}
}
