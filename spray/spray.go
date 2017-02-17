// spray is meta-package for manage possible outputs
// Basically if you want realize spray output, if realize spray interface
// After this you can use constructions like
//
// import github.com/mbobakov/spray-rabbit/spray
// import _ github.com/user/mycoolspray
//
// o,_ := spray.Get("mycoolspray")
// spr,_ := o.New(ctx,"concrete",config, nil)
// spr.Spray([]byte("Hello"))

package spray

import (
	"context"
	"sync"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/pkg/errors"
)

// Out is output factory for sprays
type Out interface {
	// Setups spray due configuration
	NewWithConfig(ctx context.Context, config ast.Node, logger Logger) (Sprayer, error)
}

// Sprayer describes spray-output
type Sprayer interface {
	//Spray message with spray
	Spray(msg []byte) error
}

// Logger describes methods for a logging process in Sprays
type Logger interface {
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
}

var (
	mutex    sync.RWMutex
	registry = make(map[string]Out)
)

// Register registers new spray output
func Register(name string, o Out) {
	mutex.Lock()
	defer mutex.Unlock()
	registry[name] = o
}

// IsRegistred check spray registration
func IsRegistred(name string) bool {
	mutex.RLock()
	defer mutex.RUnlock()
	_, ok := registry[name]
	return ok
}

// Get returns registred spray output from registry
func Get(name string) (Out, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	if !IsRegistred(name) {
		return nil, errors.Errorf("Spray '%s' not found", name)
	}
	return registry[name], nil
}

// All returns all spray-types catalog
func All() map[string]Out {
	mutex.RLock()
	defer mutex.RUnlock()
	return registry
}
