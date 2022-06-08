package app

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/homework3/comments/internal/db_model"
	"github.com/homework3/comments/internal/dto"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (a *App) getRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/comment", a.addComment)

	return r
}

func (a *App) addComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)

		return
	}

	cmt := dto.Comment{}
	err := json.NewDecoder(r.Body).Decode(&cmt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if len(cmt.Comment) == 0 || cmt.ItemId == 0 || cmt.UserId == 0 {
		http.Error(w, "json fields must be not empty", http.StatusUnprocessableEntity)
	}

	commentId, err := a.repo.AddComment(r.Context(), cmt.Comment, cmt.ItemId, cmt.UserId, db_model.CommentStatusNew)
	if err != nil {
		log.Error().Err(err).Msg("Failed to save comment into db")
		http.Error(w, "Failed to process comment", http.StatusInternalServerError)

		return
	}

	if err = a.producer.SendComment(r.Context(), db_model.Comment{
		Id:      commentId,
		UserId:  cmt.UserId,
		ItemId:  cmt.ItemId,
		Comment: cmt.Comment,
	}); err != nil {
		http.Error(w, "Failed to process comment", http.StatusInternalServerError)

		return
	}

	err = a.repo.UpdateCommentStatus(r.Context(), commentId, db_model.CommentStatusUnderModeration)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update comment status")
	}

	w.Write([]byte("Comment was added"))
}
