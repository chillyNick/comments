package mb_model

const (
	ModerationMessageStatusFailed = "failed"
	ModerationMessageStatusPassed = "passed"
)

type ModerationMessage struct {
	MessageId int
	Status    string
	Reason    string
}
