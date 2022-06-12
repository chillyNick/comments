package repository

import (
	"github.com/homework3/comments/internal/db_model"
	"golang.org/x/net/context"
)

type Repository interface {
	AddComment(ctx context.Context, comment string, itemId, userId int32, statusId string) (int64, error)
	UpdateCommentStatus(ctx context.Context, commentId int64, statusId string) error
	GetComments(ctx context.Context, itemId int32) ([]db_model.Comment, error)
}
