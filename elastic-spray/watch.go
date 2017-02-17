package elasticspray

import (
	"context"
	"time"
)

var inFlightEvents uint64

// watch watches for new events in elastic-Spray
// Events can be:
// 	1. Cancelation due context
//  2. Deadline reached for sending
//  3. New event for sending
func (e *elasticspray) watch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			e.logger.Infof("Spray is ending work")
			return
		case <-e.timer.C:
			e.sendToElasic()
			inFlightEvents = 0
		case b := <-e.sendingQueue:
			if inFlightEvents == 0 {
				e.timer.Reset(time.Duration(e.BatchDelayMs) * time.Millisecond)
			}
			err := e.addMsgToBuffer(b)
			if err != nil {
				e.logger.Errorf("indexing fail. Exiting... Err: '%s'", err)
				return
			}
			inFlightEvents++
			if inFlightEvents >= uint64(e.BatchSize) {
				e.sendToElasic()
				inFlightEvents = 0
			}
		}
	}
}
