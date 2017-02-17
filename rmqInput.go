package main

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mbobakov/spray-rabbit/spray"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type rmqInput struct {
	Ack           bool     `hcl:"ack"`
	Exchange      string   `hcl:"exchange"`
	URI           string   `hcl:"uri"`
	Key           string   `hcl:"key"`
	PrefetchCount int      `hcl:"prefetch_count"`
	Queue         string   `hcl:"queue"`
	AutoDelete    bool     `hcl:"auto_delete"`
	SpayTo        []string `hcl:"spay_to"`

	connection      *amqp.Connection
	channel         *amqp.Channel
	channelErr      chan *amqp.Error
	channelMessages <-chan amqp.Delivery
	logger          *logrus.Entry
	outputs         []spray.Sprayer
}

func getRmqs(conf *ast.ObjectList) ([]*rmqInput, error) {
	filtered := conf.Filter(rabbitmqConfKey)
	if len(filtered.Items) == 0 {
		return nil, errors.New("RabbitMQ configs for input not found")
	}
	res := make([]*rmqInput, 0)
	for _, c := range filtered.Items {
		rc := new(rmqInput)
		err := hcl.DecodeObject(&rc, c.Val)
		if err != nil {
			l.Errorf("Fail to decode config into struct. Err: '%s'. This input will be skipped.\n", err)
			continue
		}
		res = append(res, rc)
	}
	return res, nil
}

func (r *rmqInput) setup(ctx context.Context) error {
	var err error
	r.setLogger()
	r.logger.Infof("Setting up '%#v'", r)
	r.clean()

	r.connection, err = amqp.Dial(r.URI)
	if err != nil {
	connectionDone:
		for {
			select {
			case <-ctx.Done():
				return errors.New("Work ended. Connection wasn't created")
			default:
				r.logger.WithField("uri", r.URI).Errorf("Connection failed. '%s'", err)
				time.Sleep(time.Second)
				r.clean()
				r.connection, err = amqp.Dial(r.URI)
				if err == nil {
					break connectionDone
				}
			}
		}
	}
	r.logger.Infof("Connected to %s", r.URI)

	r.channel, err = r.connection.Channel()

	if err != nil {
	channelDone:
		for {
			select {
			case <-ctx.Done():
				return errors.New("Work ended. Channel wasn't created")
			default:
				r.logger.Errorf("Get channel from connection failed. '%s'", err)
				time.Sleep(time.Second)
				r.cleanChannel()
				r.channel, err = r.connection.Channel()
				if err == nil {
					break channelDone
				}
			}
		}
	}
	r.logger.Info("Channel was created successfully")
	r.channelErr = r.connection.NotifyClose(make(chan *amqp.Error, 0))
	return nil
}

func (r *rmqInput) run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	err := r.setup(ctx)
	if err != nil {
		r.logger.Errorf("Skipped. Err: '%s'", err)
		return
	}

	r.setupSprays(ctx)

	err = r.declare()
	if err != nil {
		r.logger.Errorf("Fail to start.Input was skipped. Err: '%s'", err)
	}

	for {
		select {
		case <-ctx.Done():
			r.clean()
			r.logger.Info("Work ended")
			return

		case ae := <-r.channelErr:
			r.logger.Errorf("AMQP Error. Reconnecting... Err: '%s' ", ae)
			err := r.setup(ctx)
			if err != nil {
				r.logger.Errorf("Exiting... Err: '%s' ", err)
				return
			}
			err = r.declare()
			if err != nil {
				r.logger.Errorf("Fail to redeclare. Exiting... Err: '%s' ", err)
				return
			}
		case d := <-r.channelMessages:
			for _, o := range r.outputs {
				err := o.Spray(d.Body)
				if err != nil {
					r.logger.Errorf("Fail to splay. Err: '%s'", err)
				}
			}
		}
	}
}

func (r *rmqInput) clean() {
	if r.connection != nil {
		err := r.connection.Close()
		if err != nil {
			r.logger.Errorf("Fail to close rmq-connection. Err: '%s'", err)
		}
	}
	r.cleanChannel()
}

func (r *rmqInput) cleanChannel() {
	if r.channel != nil {
		err := r.channel.Close()
		if err != nil {
			r.logger.Errorf("Fail to close rmq-channel. Err: '%s'", err)
		}
	}
}

func (r *rmqInput) setLogger() {
	r.logger = l.WithField("type", "input").WithField("name", r.Queue)
}

func (r *rmqInput) declare() error {
	err := r.channel.Qos(
		r.PrefetchCount,
		0,
		false,
	)
	if err != nil {
		return errors.Wrap(err, "Coudn't set Prefetch Count for a channel")
	}

	queue, err := r.channel.QueueDeclare(
		r.Queue,      // name of the queue
		false,        // durable
		r.AutoDelete, // delete when unused
		false,        // exclusive
		false,        // noWait
		nil,          // arguments
	)
	if err != nil {
		return errors.Wrap(err, "Coudn't declare queue")
	}
	r.logger.Infof("Queue(%s) was declared", r.Queue)
	err = r.channel.QueueBind(
		queue.Name, // name of the queue
		r.Key,      // bindingKey
		r.Exchange, // sourceExchange
		false,      // noWait
		nil,        // arguments
	)
	if err != nil {
		return errors.Wrap(err, "Coudn't  bind queue to the exchange")
	}
	r.logger.Infof("Queue('%s') was binded to exchange('%s') with the key('%s')", r.Queue, r.Exchange, r.Key)

	r.channelMessages, err = r.channel.Consume(
		r.Queue,             // name
		randStringRunes(12), // consumerTag,
		!r.Ack,              // noAck
		false,               // exclusive
		false,               // noLocal
		false,               // noWait
		nil,                 // arguments
	)
	if err != nil {
		return errors.Wrap(err, "Coudn't get channel with messages")
	}

	r.logger.Infof("Consuming for queue('%s') was started", r.Queue)

	return nil
}

func (r *rmqInput) setupSprays(ctx context.Context) {
	for _, s := range r.SpayTo {
		ns, err := sprays.get(s)
		if err != nil {
			r.logger.Errorf("Spray('%s') not found and will be skipped. Err: '%s'", s, err)
			continue
		}
		r.outputs = append(r.outputs, ns)
	}
}
