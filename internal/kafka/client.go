package kafka

import (
	"context"
	"fmt"
	"time"

	applicationMetricModel "k8s-monitoring-app/pkg/application_metric/model"
	applicationMetricValueModel "k8s-monitoring-app/pkg/application_metric_value/model"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

// CollectConsumerLag collects Kafka consumer lag information
func CollectConsumerLag(ctx context.Context, config *applicationMetricModel.Configuration) applicationMetricValueModel.MetricValue {
	if config.KafkaBootstrapServers == "" || config.KafkaConsumerGroup == "" {
		return applicationMetricValueModel.MetricValue{
			KafkaLagStatus: "error",
			KafkaError:     "missing required configuration: kafka_bootstrap_servers and kafka_consumer_group are required",
		}
	}

	// Setup Kafka client configuration
	dialer, err := createDialer(config)
	if err != nil {
		log.Error().Msg("failed to create Kafka dialer")
		return applicationMetricValueModel.MetricValue{
			KafkaLagStatus: "error",
			KafkaError:     fmt.Sprintf("failed to create Kafka dialer: %v", err),
		}
	}

	// Create Kafka client
	client := &kafka.Client{
		Addr:    kafka.TCP(config.KafkaBootstrapServers),
		Timeout: 10 * time.Second,
	}

	// Get consumer group information
	var topicsToMonitor []string
	if config.KafkaTopic != "" {
		// Monitor specific topic
		topicsToMonitor = []string{config.KafkaTopic}
	} else {
		// Get all topics for the consumer group
		conn, err := dialer.DialContext(ctx, "tcp", config.KafkaBootstrapServers)
		if err != nil {
			log.Error().Msg("failed to connect to Kafka")
			return applicationMetricValueModel.MetricValue{
				KafkaLagStatus: "error",
				KafkaError:     fmt.Sprintf("failed to connect to Kafka: %v", err),
			}
		}
		defer conn.Close()

		partitions, err := conn.ReadPartitions()
		if err != nil {
			log.Error().Msg("failed to read partitions")
			return applicationMetricValueModel.MetricValue{
				KafkaLagStatus: "error",
				KafkaError:     fmt.Sprintf("failed to read partitions: %v", err),
			}
		}

		// Extract unique topics
		topicMap := make(map[string]bool)
		for _, partition := range partitions {
			topicMap[partition.Topic] = true
		}
		for topic := range topicMap {
			topicsToMonitor = append(topicsToMonitor, topic)
		}
	}

	// Collect lag for each topic
	var topicLags []applicationMetricValueModel.KafkaTopicLag
	var totalLag int64

	for _, topic := range topicsToMonitor {
		topicLag, err := collectTopicLag(ctx, client, config, topic)
		if err != nil {
			log.Error().
				Str("topic", topic).
				Msg("failed to collect lag for topic")
			continue
		}

		topicLags = append(topicLags, topicLag)
		totalLag += topicLag.TotalLag
	}

	if len(topicLags) == 0 {
		return applicationMetricValueModel.MetricValue{
			KafkaLagStatus:     "error",
			KafkaConsumerGroup: config.KafkaConsumerGroup,
			KafkaError:         "no topics found or unable to collect lag",
		}
	}

	// Determine status based on total lag and threshold
	threshold := config.KafkaLagThreshold
	if threshold == 0 {
		threshold = 1000 // Default threshold
	}

	status := "ok"
	if totalLag > threshold*10 {
		status = "critical"
	} else if totalLag > threshold {
		status = "warning"
	}

	return applicationMetricValueModel.MetricValue{
		KafkaLagStatus:     status,
		KafkaTotalLag:      totalLag,
		KafkaConsumerGroup: config.KafkaConsumerGroup,
		KafkaTopicLags:     topicLags,
	}
}

func createDialer(config *applicationMetricModel.Configuration) (*kafka.Dialer, error) {
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	// Configure SASL if needed
	if config.KafkaSecurityProtocol != "" && config.KafkaSecurityProtocol != "PLAINTEXT" {
		var mechanism sasl.Mechanism
		var err error

		switch config.KafkaSaslMechanism {
		case "PLAIN":
			mechanism = plain.Mechanism{
				Username: config.KafkaSaslUsername,
				Password: config.KafkaSaslPassword,
			}
		case "SCRAM-SHA-256":
			mechanism, err = scram.Mechanism(scram.SHA256, config.KafkaSaslUsername, config.KafkaSaslPassword)
		case "SCRAM-SHA-512":
			mechanism, err = scram.Mechanism(scram.SHA512, config.KafkaSaslUsername, config.KafkaSaslPassword)
		default:
			return nil, fmt.Errorf("unsupported SASL mechanism: %s", config.KafkaSaslMechanism)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create SASL mechanism: %w", err)
		}

		dialer.SASLMechanism = mechanism
	}

	return dialer, nil
}

func collectTopicLag(ctx context.Context, client *kafka.Client, config *applicationMetricModel.Configuration, topic string) (applicationMetricValueModel.KafkaTopicLag, error) {
	// Get latest offsets (high water marks) first
	conn, err := kafka.Dial("tcp", config.KafkaBootstrapServers)
	if err != nil {
		return applicationMetricValueModel.KafkaTopicLag{}, fmt.Errorf("failed to dial: %w", err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions(topic)
	if err != nil {
		return applicationMetricValueModel.KafkaTopicLag{}, fmt.Errorf("failed to read partitions: %w", err)
	}

	// Build partition list for offset fetch
	partitionIDs := make([]int, len(partitions))
	for i, p := range partitions {
		partitionIDs[i] = p.ID
	}

	// Get consumer group offsets
	offsetFetchReq := &kafka.OffsetFetchRequest{
		GroupID: config.KafkaConsumerGroup,
		Topics: map[string][]int{
			topic: partitionIDs,
		},
	}

	offsetFetchResp, err := client.OffsetFetch(ctx, offsetFetchReq)
	if err != nil {
		return applicationMetricValueModel.KafkaTopicLag{}, fmt.Errorf("failed to fetch offsets: %w", err)
	}

	// Create a map of partition -> committed offset
	// offsetFetchResp.Topics is a map[string][]OffsetFetchPartition
	committedOffsets := make(map[int32]int64)
	if topicPartitions, ok := offsetFetchResp.Topics[topic]; ok {
		for _, partResp := range topicPartitions {
			committedOffsets[int32(partResp.Partition)] = partResp.CommittedOffset
		}
	}

	var partitionLags []applicationMetricValueModel.KafkaPartitionLag
	var topicTotalLag int64

	for _, partition := range partitions {
		// Get consumer offset
		currentOffset, exists := committedOffsets[int32(partition.ID)]

		// Skip if consumer hasn't committed any offset yet
		if !exists || currentOffset < 0 {
			continue
		}

		// Get log end offset using reader
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:   []string{config.KafkaBootstrapServers},
			Topic:     topic,
			Partition: partition.ID,
			MaxBytes:  1, // Read minimal data, we just need stats
		})

		// Set the reader offset to get stats
		_ = reader.SetOffset(kafka.LastOffset)
		stats := reader.Stats()
		logEndOffset := stats.Offset
		reader.Close()

		// Calculate lag
		lag := logEndOffset - currentOffset
		if lag < 0 {
			lag = 0
		}

		partitionLags = append(partitionLags, applicationMetricValueModel.KafkaPartitionLag{
			Partition:     int32(partition.ID),
			CurrentOffset: currentOffset,
			LogEndOffset:  logEndOffset,
			Lag:           lag,
		})

		topicTotalLag += lag
	}

	return applicationMetricValueModel.KafkaTopicLag{
		Topic:         topic,
		TotalLag:      topicTotalLag,
		PartitionLags: partitionLags,
	}, nil
}
