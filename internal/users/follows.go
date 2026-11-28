package users

import (
	"context"
	"errors"
	"net/http"

	"github.com/Ayushmangit/systemDesignGO/internal/platform/httpx"
	"github.com/jackc/pgx/v5"
)

// target finds the user named in the path , or answers 404
func (h *Handler) target(w http.ResponseWriter, r *http.Request) (int64, bool) {
	var id int64
	err := h.db.QueryRow(r.Context(), `select id from users where handle = $1`, r.PathValue("handle")).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.Error(w, http.StatusNotFound, "no such user")
		return 0, false
	}
	if err != nil {
		httpx.ServerError(w, r, err)
		return 0, false
	}
	return id, true
}

// Follow makes the caller follow the user in the path . Doing it twice is fine
func (h *Handler) Follow(w http.ResponseWriter, r *http.Request) {
	followee, ok := h.target(w, r)
	if !ok {
		return
	}
	if followee == ID(r.Context()) {
		httpx.Error(w, http.StatusBadRequest, "you can't follow yourself")
		return
	}
	_, err := h.db.Exec(r.Context(),
		`insert into follows (follower_id , followee_id) values($1,$2)
  on conflict do nothing`, ID(r.Context()), followee)
	if err != nil {
		httpx.ServerError(w, r, err)
		return
	}
	h.updateTimeline(r.Context(), w, r)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UnFollow(w http.ResponseWriter, r *http.Request) {
	followee, ok := h.target(w, r)
	if !ok {
		return
	}
	_, err := h.db.Exec(r.Context(),
		`delete from follows where follower_id = $1 and followee_id = $2`, ID(r.Context()), followee)
	if err != nil {
		httpx.ServerError(w, r, err)
		return
	}
	h.updateTimeline(r.Context(), w, r)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) updateTimeline(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// who they follow just changed , so their timeline is out of date.
	if err := h.tl.Drop(r.Context(), ID(r.Context())); err != nil {
		httpx.ServerError(w, r, err)
		return
	}
}
