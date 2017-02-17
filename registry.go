package main

import (
	"context"
	"sync"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/mbobakov/spray-rabbit/spray"
	"github.com/pkg/errors"
)

type sprayRegistry struct {
	sync.RWMutex
	value map[string]spray.Sprayer
}

func (r *sprayRegistry) initSprays(ctx context.Context, conf *ast.ObjectList) error {
	for stype, instance := range spray.All() {
		fl := conf.Filter(stype)
		for _, i := range fl.Items {
			for _, spr := range i.Keys {
				name := spr.Token.Value().(string)
				sl := l.WithField("type", stype).WithField("name", name)
				spray, err := instance.NewWithConfig(ctx, i.Val, sl)
				if err != nil {
					sl.Errorf("Fail to initialize. Err: '%s'", err)
					continue
				}
				err = r.add(name, spray)
				if err != nil {
					return errors.Wrap(err, "Coudn't init sprays")
				}
			}
		}
	}
	return nil
}

func (r *sprayRegistry) add(nm string, s spray.Sprayer) error {
	r.Lock()
	defer r.Unlock()
	_, ok := r.value[nm]
	if ok {
		return errors.Errorf("Spray('%s') already added", nm)
	}
	r.value[nm] = s
	return nil
}

func (r *sprayRegistry) get(nm string) (spray.Sprayer, error) {
	r.RLock()
	defer r.RUnlock()
	s, ok := r.value[nm]
	if !ok {
		return nil, errors.Errorf("Spray('%s') not found", nm)
	}
	return s, nil
}
