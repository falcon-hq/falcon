package clench

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
)

type Server struct {
	Resolve func(ctx context.Context, name string) (net.IP, error)
	Dial    func(ctx context.Context, network, addr string) (net.Conn, error)
}

func (s *Server) HandleConn(conn net.Conn) error {
	defer conn.Close()
	request, err := ReadHSRequest(conn)
	if err != nil {
		// if err == UnrecognizedAddrType {
		// 	if err := SendReply(conn, AddrTypeNotSupported, nil); err != nil {
		// 		return fmt.Errorf("Failed to send reply: %v", err)
		// 	}
		// }
		return fmt.Errorf("Failed to read destination address: %v", err)
	}

	log.Printf("opening channal to %s:%d", string(request.DST), request.PORT)

	// Process the client request
	if err := s.HandleRequest(request, conn); err != nil {
		err = fmt.Errorf("Failed to handle request: %v", err)
		log.Printf("[ERR] socks: %v", err)
		return err
	}

	return nil
}

func (s *Server) HandleRequest(req *HSRequest, conn connWriteReader) error {
	ctx := context.Background()

	// Resolve the address if we have a FQDN
	dest := req.destAddr
	if s.Resolve == nil {
		s.Resolve = func(ctx context.Context, name string) (net.IP, error) {
			addr, err := net.ResolveIPAddr("ip", name)
			if err != nil {
				return nil, err
			}
			return addr.IP, err
		}
	}

	if dest.FQDN != "" {
		addr, err := s.Resolve(ctx, dest.FQDN)
		if err != nil {
			return fmt.Errorf("Failed to resolve destination '%v': %v", dest.FQDN, err)
		}
		dest.IP = addr
	}

	return s.handleConnect(ctx, conn, req)
}

// handleConnect is used to handle a connect command
func (s *Server) handleConnect(ctx context.Context, conn connWriteReader, req *HSRequest) error {
	// Attempt to connect
	if s.Dial == nil {
		s.Dial = func(ctx context.Context, netType, addr string) (net.Conn, error) {
			return net.Dial(netType, addr)
		}
	}

	var target net.Conn
	var err error
	// TODO: add udp
	if req.PROTOCOL&0b10 == 0b10 {
		target, err = s.Dial(ctx, "udp", req.destAddr.Address())
	} else {
		target, err = s.Dial(ctx, "tcp", req.destAddr.Address())
	}

	if err != nil {
		if err := s.SendResponse(conn, []byte{1, 0}); err != nil {
			return fmt.Errorf("Failed to send reply: %v", err)
		}
		return fmt.Errorf("Connect to %v failed: %v", req.destAddr, err)
	}
	defer target.Close()

	// Send success
	if err := s.SendResponse(conn, []byte{0, 0}); err != nil {
		return fmt.Errorf("Failed to send reply: %v", err)
	}

	log.Printf("handshake done, starting channal")

	// Start proxying
	errCh := make(chan error, 2)
	go proxy(target, conn, errCh)
	go proxy(conn, target, errCh)

	log.Printf("opened channal to %s:%d", string(req.DST), req.PORT)

	// Wait
	for i := 0; i < 2; i++ {
		e := <-errCh
		if e != nil {
			// return from this function closes target (and conn).
			log.Printf("channal %s:%d close: %v", string(req.DST), req.PORT, e)
			return e
		}
	}

	return nil
}

// SendResponse is used to send a response message
func (s *Server) SendResponse(w io.Writer, statusMask []byte) error {
	// Send the message

	r := &HSResponse{
		STATUS: statusMask,
	}

	_, err := w.Write(r.Marshal())
	return err
}

type closeWriter interface {
	CloseWrite() error
}

// proxy is used to suffle data from src to destination, and sends errors
// down a dedicated channel
func proxy(dst io.Writer, src io.Reader, errCh chan error) {
	_, err := io.Copy(dst, src)
	if err != nil {
		log.Println(err)
	}

	if tcpConn, ok := dst.(closeWriter); ok {
		tcpConn.CloseWrite()
	}
	errCh <- err
}
