package health

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/codec404/chat-service/pkg/errorhandler"
	externalerror "github.com/codec404/chat-service/pkg/external_error"
	"github.com/codec404/chat-service/pkg/render"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler { return &Handler{db: db} }

// Live is the liveness probe — confirms the process is running.
// Never checks external dependencies.
func (h *Handler) Live(w http.ResponseWriter, r *http.Request) {
	render.JSONResponse(w, r, http.StatusOK, map[string]string{"status": "ok"})
}

// Ready is the readiness probe — confirms the service can serve traffic.
// Returns 503 if the database is unreachable.
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.db.Ping(r.Context()); err != nil {
		errorhandler.Respond(w, r, externalerror.New(http.StatusServiceUnavailable, "database unavailable"))
		return
	}
	render.JSONResponse(w, r, http.StatusOK, map[string]string{"status": "ready"})
}
