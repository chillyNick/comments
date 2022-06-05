package app

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/gorilla/mux"
	"github.com/homework3/comments/internal/dto"
	"github.com/homework3/comments/internal/messageBroker/kafka"
	"github.com/homework3/comments/internal/model"
	"github.com/homework3/comments/pkg/mb_model"
	"golang.org/x/net/context"
	"log"
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

	commentId, err := a.Repo.AddComment(context.Background(), cmt.Comment, cmt.ItemId, cmt.UserId, model.CommentStatusNew)
	if err != nil {
		log.Printf("failed to save comment into db: %s\n", err)
		http.Error(w, "Failed to process comment", http.StatusInternalServerError)

		return
	}

	//write to the mb
	pr := kafka.CreateProducer(a.Config.Kafka.Brokers)
	byteVal, err := json.Marshal(mb_model.Comment{
		Id:      commentId,
		UserId:  cmt.UserId,
		ItemId:  cmt.ItemId,
		Comment: cmt.Comment,
	})
	if err != nil {
		log.Printf("failed to save comment into db: %s\n", err)
		http.Error(w, "Failed to process comment", http.StatusInternalServerError)

		return
	}

	msg := sarama.ProducerMessage{
		Topic: a.Config.Kafka.ProducerTopic,
		Key:   nil,
		Value: sarama.ByteEncoder(byteVal),
	}

	_, _, err = pr.SendMessage(&msg)
	if err != nil {
		log.Printf("failed to send message into mb: %s\n", err)
		http.Error(w, "Failed to process comment", http.StatusInternalServerError)

		return
	}

	err = a.Repo.UpdateCommentStatus(context.Background(), commentId, model.CommentStatusUnderModeration)
	if err != nil {
		log.Printf("failed to update comment status: %s", err)
	}

	w.Write([]byte("Comment was added"))
}
