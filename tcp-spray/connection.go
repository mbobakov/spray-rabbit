package tcpspray

import (
	"context"
	"time"
)

func (t *tcpspray) watchConnection(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			t.logger.Infof("Stop watching connections")
			return
		case c := <-t.badConns:
			conn, err := connect(c.target, time.Duration(t.Timeout)*time.Second)
			if err != nil {
				t.logger.Errorf("Coudn't connect to %s. Reconnect in 1 sec. Err: %s", c.target, err)
				time.Sleep(time.Second)
				t.badConns <- c
				continue
			}
			t.pool <- conn
		}
	}
}
