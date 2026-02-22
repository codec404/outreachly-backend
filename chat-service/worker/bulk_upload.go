package worker

import (
	"context"

	"github.com/codec404/chat-service/pkg/conc"
	log "github.com/codec404/chat-service/pkg/logger"
)

// StartBulkUploadWorker launches the bulk-upload worker as a background goroutine.
// It stops when ctx is cancelled (graceful shutdown).
// Currently a no-op skeleton — processing logic will be added later.
func StartBulkUploadWorker(ctx context.Context) {
	conc.SafeGo("bulk-upload-worker", func() {
		log.InfofWithContext(ctx, "bulk-upload-worker: started, waiting for jobs")
		<-ctx.Done()
		log.InfofWithContext(ctx, "bulk-upload-worker: shutting down")
	})
}
