// Package tcpspray sends messages to TCP endpoints.
// If there are many servers in config value 'hosts' -
// servers will be picked in round-robin scenario
// messages delimiter is '\n' (\n in messages will be escaped)
// internal send buffer harcoded to 1024
package tcpspray

import (
	"context"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mbobakov/spray-rabbit/spray"
	"github.com/pkg/errors"
)

const (
	delimiter       = '\n'
	tcpsprayConfKey = "tcp"
)

type out struct {
}

func init() {
	spray.Register(tcpsprayConfKey, new(out))
}

type tcpspray struct {
	Hosts      []string `hcl:"hosts"`
	Timeout    int      `hcl:"timeout_sec"`
	Only       *only    `hcl:"only"`
	Buffer     int      `hcl:"buffer"`
	sendBuffer chan []byte
	logger     spray.Logger
	pool       chan *connector
	badConns   chan *connector
}

type only struct {
	Exist string `hcl:"exist"`
}

func (o *out) NewWithConfig(ctx context.Context, conf ast.Node, l spray.Logger) (spray.Sprayer, error) {
	t := &tcpspray{
		logger: new(discardLog),
	}
	if l != nil {
		t.logger = l
	}
	err := t.configure(conf)
	if err != nil {
		return nil, errors.Wrap(err, "Coudn't configure tcp-spay")
	}
	err = t.validate()
	if err != nil {
		return nil, errors.Wrap(err, "tcp-spray configuration failed")
	}
	t.logger.Infof("Starting tcp-spray")
	t.sendBuffer = make(chan []byte, t.Buffer)
	t.pool = make(chan *connector, len(t.Hosts))
	t.badConns = make(chan *connector, len(t.Hosts))
	for _, s := range t.Hosts {
		t.badConns <- &connector{target: s}
	}
	go t.watchConnection(ctx)
	go t.send(ctx)
	return t, nil
}

func (t *tcpspray) configure(conf ast.Node) error {
	err := hcl.DecodeObject(t, conf)
	if err != nil {
		return errors.Wrap(err, "Coudn't decode tcp-spray configuration")
	}
	return nil
}

func (t *tcpspray) validate() error {
	if len(t.Hosts) == 0 {
		return errors.New("tcp.hosts is required parameter. Example: host = [\"server1:1111\", \"server2:2222\"]")
	}
	if t.Timeout <= 0 {
		return errors.New("tcp.timeout_sec must be positive integer")
	}
	if t.Buffer <= 0 {
		return errors.New("tcp.buffer must be positive integer")
	}
	return nil
}

type discardLog struct{}

func (d *discardLog) Errorf(format string, args ...interface{}) {}
func (d *discardLog) Infof(format string, args ...interface{})  {}
