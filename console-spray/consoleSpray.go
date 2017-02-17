package consolespray

import (
	"context"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mbobakov/spray-rabbit/spray"
)

const consprayConfKey = "console"

func init() {
	spray.Register(consprayConfKey, new(out))
}

type out struct{}

type conspray struct {
	logger spray.Logger
}

// NewWithConfig creates new instance for a console output
func (o *out) NewWithConfig(ctx context.Context, conf ast.Node, l spray.Logger) (spray.Sprayer, error) {
	c := &conspray{logger: new(discardLog)}
	if l != nil {
		c.logger = l
	}
	return c, nil
}

func (c *conspray) Spray(msg []byte) error {
	c.logger.Infof("\n---\n%s\n---\n", msg)
	return nil
}

type discardLog struct{}

func (d *discardLog) Errorf(format string, args ...interface{}) {}
func (d *discardLog) Infof(format string, args ...interface{})  {}
