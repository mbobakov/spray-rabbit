package elasticspray

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// indexContext is context of variables for create elasticsearch index name
// It is counted for every message. If it will be botleneck, please create issue
type indexContext struct {
	Year  int
	Month int
	Day   int
}

// sendToElasic is main function for iteract with Elastic search cluster
func (e *elasticspray) sendToElasic() {
	// Stop ticker
	e.timer.Stop()

	conn := <-e.pool
	defer func() { e.pool <- conn }()

	r, err := http.Post(conn, "text/plain", e.buffer)
	if err != nil {
		e.logger.Errorf("Fail to send bucket because '%s'", err)
		return
	}
	defer e.buffer.Reset()

	bd, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		e.logger.Errorf("Not successful elastic operation. Code=%d. Response='%s'", r.StatusCode, bd)
	}
}

// addMsgToBuffer wraps message for elasticSerch bulk-API mode
// See more: https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-bulk.html
func (e *elasticspray) addMsgToBuffer(msg []byte) error {
	t := time.Now()
	c := &indexContext{
		Year:  t.Year(),
		Month: int(t.Month()),
		Day:   t.Day(),
	}
	e.buffer.WriteString(`{"index":{"_id":null,"_index":"`)
	err := e.indexTemplate.Execute(e.buffer, c)
	if err != nil {
		return errors.Wrap(err, "Coudn't execute index template")
	}
	e.buffer.WriteString(fmt.Sprintf(`","_type":"%s",`, e.Type))
	e.buffer.WriteString("\"_routing\":null}}\n")
	e.buffer.Write(msg)
	e.buffer.WriteString("\n")
	return nil
}
