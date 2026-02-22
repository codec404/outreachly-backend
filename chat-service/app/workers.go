package app

import (
	"context"
	"fmt"

	"github.com/codec404/chat-service/worker"
)

const (
	BulkUploadWorker = "bulk-upload-worker"
)

// workerFuncs maps each worker name to its start function.
// Add new workers here as they are introduced.
// The map key is the exact value expected in the RUN_MODE env var.
var workerFuncs = map[string]func(context.Context){
	BulkUploadWorker: worker.StartBulkUploadWorker,
}

// RunWorker starts the named worker and blocks until ctx is cancelled.
// Returns an error immediately if the name is not registered — this surfaces
// misconfigured ECS task definitions / docker-compose env vars at startup.
func RunWorker(ctx context.Context, name string) error {
	start, ok := workerFuncs[name]
	if !ok {
		return fmt.Errorf("unknown worker %q: valid values are %v", name, validWorkerNames())
	}
	start(ctx)
	<-ctx.Done()
	return nil
}

func validWorkerNames() []string {
	names := make([]string, 0, len(workerFuncs))
	for name := range workerFuncs {
		names = append(names, name)
	}
	return names
}
