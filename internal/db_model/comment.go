package db_model

const (
	CommentStatusNew              = "new"
	CommentStatusUnderModeration  = "under_moderation"
	CommentStatusModerationFailed = "moderation_failed"
	CommentStatusModerationPassed = "moderation_passed"
)

type Comment struct {
	Id       int64
	UserId   int32
	ItemId   int32
	Comment  string
	StatusId string
}
