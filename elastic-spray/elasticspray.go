// Package elasticspray is spray realization
// This package sends messages to ElasticSearch
package elasticspray

import (
	"bytes"
	"context"
	"html/template"
	"time"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mbobakov/spray-rabbit/spray"
	"github.com/pkg/errors"
)

func init() {
	spray.Register(elasticConfKey, new(out))
}

const (
	elasticConfKey      = "elastic"
	elasticAPIBulkPath  = "_bulk"
	defaultElasticSheme = "http"
)

var warningShowed bool

type out struct{}

type elasticspray struct {
	Hosts        []string `hcl:"hosts"`
	BatchSize    int      `hcl:"batch_size"`
	BatchDelayMs int      `hcl:"batch_delay_ms"`
	Index        string   `hcl:"index"`
	Type         string   `hcl:"type"`

	indexTemplate *template.Template
	pool          chan string
	logger        spray.Logger
	sendingQueue  chan []byte
	timer         *time.Timer
	buffer        *bytes.Buffer
}

// Spray add message into sending queue
// Realize spay.Sprayer interface
func (e *elasticspray) Spray(msg []byte) error {
	select {
	case e.sendingQueue <- msg:
		warningShowed = false
	default:
		if !warningShowed {
			e.logger.Errorf("All next messages will be dropped because sending buffer is full")
			warningShowed = true
		}
	}
	return nil
}

// NewWithConfig creates new instance of elastic spray with configuration in config
// <conf> parameter will try to decode into <elasticspray> struct
// Realize spray.Out interface
func (o *out) NewWithConfig(ctx context.Context, conf ast.Node, l spray.Logger) (spray.Sprayer, error) {
	e := &elasticspray{
		logger: new(discardLog),
	}
	if l != nil {
		e.logger = l
	}

	err := e.configure(conf)
	if err != nil {
		return nil, errors.Wrap(err, "Coudn't configure for elastic-spay")
	}
	err = e.validate()
	if err != nil {
		return nil, errors.Wrap(err, "elastic-spray configuration failed")
	}
	e.logger.Infof("Starting elastic-spray")

	e.pool = make(chan string, len(e.Hosts))
	for _, s := range e.Hosts {
		h, err1 := normalizeURL(s)
		if err1 != nil {
			e.logger.Errorf("Target endpoint('%s') will be skipped. Err: %s", s, err)
			continue
		}
		e.pool <- h
	}

	if len(e.pool) == 0 {
		return nil, errors.New("No endpoints for sending. elastic-spray is exiting")
	}

	e.sendingQueue = make(chan []byte, 1024*len(e.pool))
	e.timer = time.NewTimer(time.Duration(e.BatchDelayMs) * time.Millisecond)
	e.timer.Stop()
	e.buffer = bytes.NewBuffer(nil)
	e.indexTemplate, err = template.New("indexname").Parse(e.Index)
	if err != nil {
		return nil, errors.Wrap(err, "Index template is not valid")
	}
	go e.watch(ctx)
	return e, nil
}

func (e *elasticspray) configure(conf ast.Node) error {
	err := hcl.DecodeObject(e, conf)
	if err != nil {
		return errors.Wrap(err, "Coudn't decode elastic-spray configuration")
	}
	return nil
}

func (e *elasticspray) validate() error {
	if len(e.Hosts) == 0 {
		return errors.New("Please specify elastic.hosts correctly. Not hosts found")
	}
	if e.BatchSize <= 0 {
		return errors.New("Please specify elastic.batch_size correctly. This must be greater than 0")
	}
	if e.BatchDelayMs <= 0 {
		return errors.New("Please specify elastic.batch_delay_ms correctly. This must be greater than 0")
	}
	if len(e.Type) == 0 {
		return errors.New("Please specify elastic.type correctly. This must be valid string")
	}
	return nil
}

type discardLog struct{}

func (d *discardLog) Errorf(format string, args ...interface{}) {}
func (d *discardLog) Infof(format string, args ...interface{})  {}
