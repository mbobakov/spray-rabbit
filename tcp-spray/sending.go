package tcpspray

import (
	"bytes"
	"context"
	"time"

	"github.com/buger/jsonparser"
)

var warningShowed bool

func (t *tcpspray) Spray(msg []byte) error {
	if t.pass(msg) {
		select {
		case t.sendBuffer <- wrap(msg):
			warningShowed = false
		default:
			if !warningShowed {
				t.logger.Errorf("All next messages will be dropped because sending buffer is full")
				warningShowed = true
			}
		}
	}
	return nil
}

//
func (t *tcpspray) pass(msg []byte) bool {
	if t.Only == nil {
		return true
	}
	if t.Only.Exist == "" {
		return true
	}
	_, _, _, err := jsonparser.Get(msg, t.Only.Exist)
	return err != nil
}

// wrap escaped delimiter char in a message and add delimiter at the end
func wrap(src []byte) []byte {
	w := bytes.Replace(src, []byte{delimiter}, append([]byte("\\"), delimiter), -1)
	return append(w, delimiter)
}

func (t *tcpspray) send(ctx context.Context) {
	t.logger.Infof("Sending was started")
	for {
		select {
		case <-ctx.Done():
			t.logger.Infof("Spray is ending work")
			return
		case msg := <-t.sendBuffer:
			conn := <-t.pool
			conn.conn.SetDeadline(time.Now().Add(time.Duration(t.Timeout) * time.Second))
			_, err := conn.conn.Write(msg)
			if err != nil {
				t.logger.Errorf("Fail to write data into socket '%s' . Err '%s'", conn.target, err)
				t.badConns <- conn
				break
			}
			t.pool <- conn
		}
	}
}
