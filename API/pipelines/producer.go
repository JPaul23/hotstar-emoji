package pipelines

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
)

// set up producer
func KafkaProducerSetup() *kafka.Writer {
	err := godotenv.Load()
	if err != nil {
		panic("====> Error in loading api .env file")
	}
	kafkaHost := os.Getenv("KAKFA_HOST")
	kafkaPort := os.Getenv("KAFKA_PORT")
	emojiKafkaTopic := os.Getenv("KAFKA_TOPIC")

	writer := kafka.NewWriter(kafka.WriterConfig{
		// Brokers: []string{"%s:%s", kafkaHost, kafkaPort},
		Brokers:  []string{fmt.Sprintf("%s:%s", kafkaHost, kafkaPort)},
		Topic:    emojiKafkaTopic,
		Balancer: &kafka.LeastBytes{},
	})

	return writer
}

// kafka producer func
func KafkaMessageProducer(writer *kafka.Writer, key, value []byte) error {
	return writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   key,
			Value: value,
		})
}
