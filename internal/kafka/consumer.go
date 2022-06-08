package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/homework3/comments/internal/config"
	"github.com/homework3/comments/internal/db_model"
	"github.com/homework3/comments/internal/repository"
	"github.com/homework3/comments/internal/tracer"
	"github.com/homework3/comments/pkg/model"
	"github.com/opentracing/opentracing-go"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
)

type consumerHandler struct {
	repo repository.Repository
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *consumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	log.Info().Msg("Setup consumer group session")

	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	log.Info().Msg("cleanup")

	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.Info().Msg(fmt.Sprintf("Start consumer loop for topic: %s", claim.Topic()))
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// <https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29>
	for msg := range claim.Messages() {
		log.Info().
			Str("value", string(msg.Value)).
			Msgf("Message topic:%q partition:%d offset:%d", msg.Topic, msg.Partition, msg.Offset)

		spanCtx, err := tracer.ExtractSpanContext(msg.Headers)
		if err != nil {
			log.Error().Err(err).Msg("Failed to extract spanContext from kafka consumer headers")
		}
		span := opentracing.StartSpan("Comment after moderation processing", opentracing.ChildOf(spanCtx))

		comment := model.ModerationComment{}
		err = json.Unmarshal(msg.Value, &comment)
		if err != nil {
			log.Error().Err(err).Str("value", string(msg.Value)).Msg("Failed to unmarshal comment")
			span.Finish()

			continue
		}

		var status int
		switch comment.Status {
		case model.ModerationCommentStatusPassed:
			status = db_model.CommentStatusModerationPassed
		case model.ModerationCommentStatusFailed:
			status = db_model.CommentStatusModerationFailed
		default:
			log.Error().Msgf("Failed to define comment status: %s", comment.Status)
		}

		err = c.repo.UpdateCommentStatus(session.Context(), comment.CommentId, status)
		if err != nil {
			log.Error().Err(err).Msg("Failed to update comment status")
			span.Finish()

			continue
		}

		session.MarkMessage(msg, "")
		span.Finish()
	}

	return nil
}

func StartProcessMessages(ctx context.Context, repo repository.Repository, cfg *config.Kafka) error {
	consumer, err := createConsumerGroup(cfg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create a consumer group")

		return err
	}
	defer consumer.Close()

	handler := &consumerHandler{repo: repo}
	loop := true
	for loop {
		err = consumer.Consume(ctx, []string{cfg.ConsumerTopic}, handler)
		if err != nil {
			log.Error().Err(err).Msg(" Consumer group session error")
		}

		select {
		case <-ctx.Done():
			loop = false
		default:

		}
	}

	return nil
}

func createConsumerGroup(cfg *config.Kafka) (sarama.ConsumerGroup, error) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	return sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupId, saramaCfg)
}
