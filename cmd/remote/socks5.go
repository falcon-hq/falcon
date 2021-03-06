package remotecmd

import (
	"fmt"
	"net"

	"github.com/maoxs2/go-socks5"
)

type s5RemoteServer struct {
	*socks5.Server
}

// overwrite the serverConn
func (s *s5RemoteServer) serverConn(conn net.Conn) error {
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
		err = fmt.Errorf("failed to handle request: %v", err)
		log.Errorf("socks error: %v", err)
		return err
	}

	return nil
}
