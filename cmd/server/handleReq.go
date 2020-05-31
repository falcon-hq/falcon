package main

import (
	"fmt"
	"net"

	"github.com/maoxs2/go-socks5"
)

type server struct {
	*socks5.Server
}

// overwrite the serverConn
func (s *server) handleConn(conn net.Conn) error {
	defer conn.Close()

	request, err := socks5.NewRequest(conn)
	if err != nil {
		if err == socks5.UnrecognizedAddrType {
			if err := socks5.SendReply(conn, socks5.AddrTypeNotSupported, nil); err != nil {
				return fmt.Errorf("Failed to send reply: %v", err)
			}
		}
		return fmt.Errorf("Failed to read destination address: %v", err)
	}

	if client, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		request.RemoteAddr = &socks5.AddrSpec{IP: client.IP, Port: client.Port}
	}

	// Process the client request
	if err := s.HandleRequest(request, conn); err != nil {
		err = fmt.Errorf("Failed to handle request: %v", err)
		s.Config.Logger.Printf("[ERR] socks: %v", err)
		return err
	}

	return nil
}
