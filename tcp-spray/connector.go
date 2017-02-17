package tcpspray

import (
	"net"
	"time"

	"github.com/pkg/errors"
)

type connector struct {
	target string
	conn   net.Conn
}

func connect(remote string, timeout time.Duration) (*connector, error) {
	conn, err := net.DialTimeout("tcp", remote, timeout)
	if err != nil {
		return nil, errors.Wrapf(err, "Coudn't connect to %s.", remote)
	}
	return &connector{target: remote, conn: conn}, nil
}
