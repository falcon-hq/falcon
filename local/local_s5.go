package local

import (
	"fmt"
	"net"

	"github.com/maoxs2/go-socks5"
)

type s5LocalServer struct {
	*socks5.Server // act as auth
}

func (s *s5LocalServer) HandleAuth(conn net.Conn) error {
	// Read the version byte
	version := []byte{0}
	if _, err := conn.Read(version); err != nil {
		s.Config.Logger.Printf("[ERR] socks: Failed to get version byte: %v", err)
		return err
	}

	// Ensure we are compatible
	if version[0] != socks5.Socks5Version {
		err := fmt.Errorf("Unsupported SOCKS version: %v", version)
		s.Config.Logger.Printf("[ERR] socks: %v", err)
		return err
	}

	// Authenticate the connection
	_, err := s.Authenticate(conn, conn)
	if err != nil {
		err = fmt.Errorf("Failed to authenticate: %v", err)
		s.Config.Logger.Printf("[ERR] socks: %v", err)
		return err
	}

	return nil
}
