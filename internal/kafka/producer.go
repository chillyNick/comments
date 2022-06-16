package kafka

import (
	"encoding/json"

	"github.com/Shopify/sarama"
	"github.com/homework3/comments/internal/config"
	"github.com/homework3/comments/internal/db_model"
	"github.com/homework3/comments/internal/tracer"
	"github.com/opentracing/opentracing-go"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
)

type Producer struct {
	sarama.SyncProducer
	producerTopic string
}

func CreateProducer(cfg *config.Kafka) (producer Producer, err error) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.RequiredAcks = sarama.WaitForAll
	saramaCfg.Producer.Retry.Max = 10
	saramaCfg.Producer.Return.Successes = true

	syncProducer, err := sarama.NewSyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		return producer, err
	}
	producer.SyncProducer = syncProducer
	producer.producerTopic = cfg.ProducerTopic

	return producer, nil
}

func (k Producer) SendComment(ctx context.Context, comment db_model.Comment) error {
	byteVal, err := json.Marshal(comment)
	if err != nil {
		log.Error().Err(err).Msg("Failed marshalize comment")

		return err
	}

	headers := make([]sarama.RecordHeader, 0, 1)
	if span := opentracing.SpanFromContext(ctx); span != nil {
		if err = tracer.InjectSpanContext(span.Context(), &headers); err != nil {
			log.Error().Err(err).Msg("Failed to inject span contex into kafka header")
		}
	} else {
		log.Info().Msg("Failed to extract span from contex")
	}

	_, _, err = k.SendMessage(&sarama.ProducerMessage{
		Topic:   k.producerTopic,
		Value:   sarama.ByteEncoder(byteVal),
		Headers: headers,
	})
	if err != nil {
		log.Printf("Failed to send message into mb: %s\n", err)

		return err
	}

	return nil
}
