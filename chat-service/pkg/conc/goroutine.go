// Package conc provides safe concurrency helpers built on top of sourcegraph/conc.
package conc

import (
	"github.com/sourcegraph/conc/panics"

	log "github.com/codec404/chat-service/pkg/logger"
)

// SafeGo spawns f in a new goroutine and provides:
//   - Panic recovery: any unhandled panic is logged; the process does not crash.
//   - Start/stop observability so goroutine lifecycle is visible in logs.
//
// For long-running goroutines with a work loop, use SafeTry per iteration so a
// single failing iteration does not exit the loop.
func SafeGo(name string, f func()) {
	go func() {
		log.Infof("goroutine[%s] started", name)
		defer log.Infof("goroutine[%s] stopped", name)

		if rec := panics.Try(f); rec != nil {
			log.Errorf("goroutine[%s] panic: %s", name, rec.AsError())
		}
	}()
}

// SafeTry executes f synchronously, recovering any panic and logging it with
// the goroutine name. Use this inside a SafeGo work loop so a single failing
// iteration does not terminate the loop.
func SafeTry(name string, f func()) {
	if rec := panics.Try(f); rec != nil {
		log.Errorf("goroutine[%s] iteration panic: %s", name, rec.AsError())
	}
}
