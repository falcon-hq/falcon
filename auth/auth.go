package auth

import (
	"fmt"
	"io"
)

const (
	Socks5Version   = byte(0x05)
	NoAuth          = byte(0)
	NoAcceptable    = byte(0xFF)
	UserPassAuth    = byte(0x02)
	UserAuthVersion = byte(0x01)
	AuthSuccess     = byte(0)
	AuthFailure     = byte(0x01)
)

type Authenticator struct {
	userPass map[string]string
}

func NewAuthenticator(userPass map[string]string) *Authenticator {
	return &Authenticator{userPass: userPass}
}

func (a *Authenticator) Auth(reader io.Reader, writer io.Writer) error {
	if a.userPass == nil {
		_, err := writer.Write([]byte{Socks5Version, NoAuth})
		if err != nil {
			return err
		}

		header := []byte{0, 0}
		if _, err := io.ReadAtLeast(reader, header, 2); err != nil {
			return err
		}

		return nil
	}

	// Tell the client to use user/pass auth
	if _, err := writer.Write([]byte{Socks5Version, UserPassAuth}); err != nil {
		return err
	}

	// Get the version and username length
	header := []byte{0, 0}
	if _, err := io.ReadAtLeast(reader, header, 2); err != nil {
		return err
	}

	// Ensure we are compatible
	if header[0] != UserAuthVersion {
		return fmt.Errorf("unsupported auth version: %v", header[0])
	}

	// Get the user name
	userLen := int(header[1])
	user := make([]byte, userLen)
	if _, err := io.ReadAtLeast(reader, user, userLen); err != nil {
		return err
	}

	// Get the password length
	if _, err := reader.Read(header[:1]); err != nil {
		return err
	}

	// Get the password
	passLen := int(header[0])
	pass := make([]byte, passLen)
	if _, err := io.ReadAtLeast(reader, pass, passLen); err != nil {
		return err
	}

	// Verify the password
	if a.userPass[string(user)] == string(pass) {
		if _, err := writer.Write([]byte{UserAuthVersion, AuthSuccess}); err != nil {
			return err
		}
	} else {
		if _, err := writer.Write([]byte{UserAuthVersion, AuthFailure}); err != nil {
			return err
		}
		return fmt.Errorf("user auth failed")
	}

	return nil
}
