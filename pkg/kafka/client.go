package kafka

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"time"
)

type KafkaSettings struct {
	Network       string
	ListenAddress string
	Topic         string
	Partition     int
}

type KafkaClient struct {
	Client *kafka.Conn
}

type KafkaMessage struct {
	Message string `json:"message"`
}

func NewKafkaWriter(settings KafkaSettings) (*KafkaClient, error) {
	log.Info().Msgf("connection kafka settings: %v", settings)
	conn, err := kafka.DialLeader(context.Background(), settings.Network, settings.ListenAddress, settings.Topic, settings.Partition)
	if err != nil {
		log.Error().Err(err).Msg("kafka errors")
		return nil, err
	}
	err = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		log.Error().Err(err).Msg("kafka SetWriteDeadline errors")
		return nil, err
	}

	return &KafkaClient{
		Client: conn,
	}, nil
}
