package main

// If you want more spray realization include them here
import (
	_ "github.com/mbobakov/spray-rabbit/console-spray"
	_ "github.com/mbobakov/spray-rabbit/elastic-spray"
	"github.com/mbobakov/spray-rabbit/spray"
	_ "github.com/mbobakov/spray-rabbit/tcp-spray"
)

var sprays = &sprayRegistry{value: make(map[string]spray.Sprayer)}
