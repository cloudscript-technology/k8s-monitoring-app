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
	if config.KafkaBootstrapServers == "" {
		return applicationMetricValueModel.MetricValue{
			KafkaLagStatus: "error",
			KafkaError:     "missing required configuration: kafka_bootstrap_servers is required",
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

	// Create Kafka client, wiring SASL via Transport when configured
	var transport *kafka.Transport
	if dialer.SASLMechanism != nil {
		transport = &kafka.Transport{SASL: dialer.SASLMechanism}
	}
	client := &kafka.Client{
		Addr:      kafka.TCP(config.KafkaBootstrapServers),
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	// Determine topics to monitor (specific or all)
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

	// Determine consumer groups to monitor (specific or all)
	var groupsToMonitor []string
	if config.KafkaConsumerGroup != "" {
		groupsToMonitor = []string{config.KafkaConsumerGroup}
	} else {
		// List all consumer groups
		listResp, err := client.ListGroups(ctx, &kafka.ListGroupsRequest{})
		if err != nil {
			log.Error().Msg("failed to list consumer groups")
			return applicationMetricValueModel.MetricValue{
				KafkaLagStatus: "error",
				KafkaError:     fmt.Sprintf("failed to list consumer groups: %v", err),
			}
		}
		for _, g := range listResp.Groups {
			if g.GroupID != "" {
				groupsToMonitor = append(groupsToMonitor, g.GroupID)
			}
		}
		if len(groupsToMonitor) == 0 {
			return applicationMetricValueModel.MetricValue{
				KafkaLagStatus: "error",
				KafkaError:     "no consumer groups found",
			}
		}
	}

	var groupLags []applicationMetricValueModel.KafkaGroupLag
	var overallTotalLag int64

	// Collect lag per group and per topic
	for _, group := range groupsToMonitor {
		var topicLags []applicationMetricValueModel.KafkaTopicLag
		var totalLag int64

		for _, topic := range topicsToMonitor {
			topicLag, err := collectTopicLag(ctx, client, config, group, topic, dialer)
			if err != nil {
				log.Error().Str("group", group).Str("topic", topic).Msg("failed to collect lag for topic")
				continue
			}
			topicLags = append(topicLags, topicLag)
			totalLag += topicLag.TotalLag
		}

		if totalLag > 0 || len(topicLags) > 0 {
			groupLags = append(groupLags, applicationMetricValueModel.KafkaGroupLag{
				Group:     group,
				TotalLag:  totalLag,
				TopicLags: topicLags,
			})
			overallTotalLag += totalLag
		}
	}

	if len(groupLags) == 0 {
		return applicationMetricValueModel.MetricValue{
			KafkaLagStatus: "error",
			KafkaError:     "no topics/groups found or unable to collect lag",
		}
	}

	// Determine overall status based on sum of lag and threshold
	threshold := config.KafkaLagThreshold
	if threshold == 0 {
		threshold = 1000 // Default threshold
	}

	status := "ok"
	if overallTotalLag > threshold*10 {
		status = "critical"
	} else if overallTotalLag > threshold {
		status = "warning"
	}

	return applicationMetricValueModel.MetricValue{
		KafkaLagStatus: status,
		KafkaTotalLag:  overallTotalLag,
		// KafkaConsumerGroup left empty when listing all groups
		KafkaGroupLags: groupLags,
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

func collectTopicLag(ctx context.Context, client *kafka.Client, config *applicationMetricModel.Configuration, group string, topic string, dialer *kafka.Dialer) (applicationMetricValueModel.KafkaTopicLag, error) {
	// Get latest offsets (high water marks) first
	conn, err := dialer.DialContext(ctx, "tcp", config.KafkaBootstrapServers)
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
		GroupID: group,
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

		// Get log end offset using a leader connection for accuracy
		leaderConn, err := dialer.DialLeader(ctx, "tcp", config.KafkaBootstrapServers, topic, partition.ID)
		if err != nil {
			// If we can't dial leader, skip this partition but log the error
			log.Error().Str("topic", topic).Int("partition", partition.ID).Msg("failed to dial leader for partition")
			continue
		}
		logEndOffset, err := leaderConn.ReadLastOffset()
		leaderConn.Close()
		if err != nil {
			log.Error().Str("topic", topic).Int("partition", partition.ID).Msg("failed to read last offset for partition")
			continue
		}

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
