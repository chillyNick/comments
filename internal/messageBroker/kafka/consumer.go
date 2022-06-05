package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/homework3/comments/internal/config"
	"github.com/homework3/comments/internal/repository"
	"golang.org/x/net/context"
	"log"
)

type CounsumerHandler struct {
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (ch *CounsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	log.Print("setup")
	log.Print(session.Claims())
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (ch *CounsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	log.Print("cleanup")
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (ch *CounsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// <https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29>
	// Specific consumption news
	for message := range claim.Messages() {
		log.Print("[topic:%s] [partiton:%d] [offset:%d] [value:%s] [time:%v]",
			message.Topic, message.Partition, message.Offset, string(message.Value), message.Timestamp)
		// Update displacement
		session.MarkMessage(message, "")
	}

	return nil
}

func ObserveMbMessage(cnf *config.Kafka, repo repository.Repository) {
	cns, err := sarama.NewConsumerGroup(cnf.Brokers, cnf.GroupId, nil)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}
	defer cns.Close()

	ctx := context.Background()
	for {
		handler := new(CounsumerHandler)
		err := cns.Consume(ctx, []string{cnf.GroupId}, handler)
		if err != nil {
			log.Println(err)
		}
	}
}
