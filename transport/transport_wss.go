package transport

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"github.com/pilafusama/gossip/base"
)

type Wss struct {
	Ws
	certPath, keyPath string
}

func NewWss(output chan base.SipMessage, certPath, keyPath string) (*Wss, error) {
	w := Wss{}
	w.certPath = certPath
	w.keyPath = keyPath
	w.network = "wss"
	w.output = output
	w.listeningPoints = make([]*net.TCPListener, 0)
	w.connTable.Init()
	w.dialer.Protocols = []string{wsSubProtocol}
	w.dialer.Timeout = time.Minute
	w.dialer.TLSConfig = &tls.Config{
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			return nil
		},
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

	l, err := tls.Listen("tcp", addr.String(), &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	if err != nil {
		return err
	}

	lp := l.(*net.TCPListener)
	w.listeningPoints = append(w.listeningPoints, lp)
	go w.serve(lp)

	// At this point, err should be nil but let's be defensive.
	return err
}
