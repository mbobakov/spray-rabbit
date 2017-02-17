package main

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/sirupsen/logrus"
)

const rabbitmqConfKey = "rabbitmq"

var (
	configPath   = flag.String("config", "./spray-rabbit.hcl", "Path to the configuration")
	l            = logrus.New()
	globalConfig *ast.ObjectList
)

func main() {
	flag.Parse()
	l.Level = logrus.InfoLevel

	confBytes, err := ioutil.ReadFile(*configPath)
	if err != nil {
		l.Fatalf("Coudn't read configuration from %s", *configPath)
	}
	globalConfig, err = loadConfiguration(bytes.NewBuffer(confBytes))
	if err != nil {
		l.Fatalln("Coudn't load configuration")
	}
	run()
}

func run() {
	wg := new(sync.WaitGroup)
	inputs, err := getRmqs(globalConfig)
	if err != nil {
		l.Errorf("Coudn't start inputs. Err: '%s'", err)
	}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	err = sprays.initSprays(ctx, globalConfig)
	if err != nil {
		l.Fatalf("Fail to init sprays. Err: '%s'", err)
	}
	for _, in := range inputs {
		wg.Add(1)
		go in.run(ctx, wg)
	}

	go cancelIfSIGINTR(cancel)
	wg.Wait()
}

func cancelIfSIGINTR(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	l.Infoln("SIGINT Retrieved... Gracefully shutdown... ")
	cancel()
}
