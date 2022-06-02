package app

import (
	"encoding/json"
	"github.com/homework3/comments/internal/dto"
	"github.com/homework3/comments/internal/messageBroker/kafka"
	"github.com/homework3/comments/internal/model"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"net/http/pprof"
)

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/comment", a.addComment)
	attachPprof(mux)

	return mux
}

func attachPprof(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
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

	commentId, err := a.Repo.AddComment(context.Background(), cmt.Comment, cmt.ItemId, cmt.UserId, model.CommentStatusNew)
	if err != nil {
		log.Printf("failed to save comment into db: %s\n", err)
		http.Error(w, "Failed to process comment", http.StatusInternalServerError)

		return
	}

	//write to the mb
	kafka.CreateProducer(a.Config.Kafka.Brokers)

	err = a.Repo.UpdateCommentStatus(context.Background(), commentId, model.CommentStatusUnderModeration)
	if err != nil {
		log.Printf("failed to update comment status: %s", err)
	}

	w.Write([]byte("Comment was added"))
}
