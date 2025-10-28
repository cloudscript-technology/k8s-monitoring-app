-- Remove KafkaConsumerLag metric type
DELETE FROM metric_types WHERE name = 'KafkaConsumerLag';

