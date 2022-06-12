package repository

import "golang.org/x/net/context"

type Repository interface {
	AddComment(ctx context.Context, comment string, itemId, userId int32, statusId string) (int64, error)
	UpdateCommentStatus(ctx context.Context, commentId int64, statusId string) error
}
