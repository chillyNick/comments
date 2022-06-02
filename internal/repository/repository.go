package repository

import "golang.org/x/net/context"

type Repository interface {
	AddComment(ctx context.Context, comment string, itemId, userId int32, statusId int) (int, error)
	UpdateCommentStatus(ctx context.Context, commentId, statusId int) error
}
