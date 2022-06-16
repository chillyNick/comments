package app

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/homework3/comments/internal/db_model"
	"github.com/homework3/comments/internal/dto"
	"github.com/rs/zerolog/log"
)

func (a *App) getRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/comments", a.getComments)
	r.HandleFunc("/comment", a.addComment)

	return r
}

func (a *App) addComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

func (a *App) getComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)

		return
	}

	if r.URL.Query().Get("itemId") == "" {
		http.Error(w, "itemId must be set", http.StatusBadRequest)

		return
	}

	itemId, err := strconv.Atoi(r.URL.Query().Get("itemId"))
	if err != nil || itemId < 0 {
		http.Error(w, "itemId must be a positive number", http.StatusBadRequest)

		return
	}

	commentsBytes, err := a.cache.GetComments(r.Context(), int32(itemId))
	if err != nil {
		comments, err := a.repo.GetComments(r.Context(), int32(itemId))
		if err != nil {
			log.Error().Err(err).Msg("Failed to get comments from db")
			http.Error(w, "Failed to get comments", http.StatusInternalServerError)

			return
		}

		if commentsBytes, err = json.Marshal(comments); err != nil {
			log.Error().Err(err).Msg("Failed to marshalize comments")
			http.Error(w, "Failed to get comments", http.StatusInternalServerError)

			return
		}

		a.cache.SetComments(r.Context(), int32(itemId), commentsBytes)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(commentsBytes)
}
