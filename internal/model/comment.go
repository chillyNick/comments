package model

const (
	CommentStatusNew = iota + 1
	CommentStatusUnderModeration
	CommentStatusModerationFailed
	CommentStatusModerationPassed
)

type Comment struct {
	Id       int64
	UserId   int32
	ItemId   int32
	Comment  string
	StatusId int32
}
