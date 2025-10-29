-- Add KafkaConsumerLag metric type
INSERT INTO metric_types (name, description) VALUES 
('KafkaConsumerLag', 'Monitor Kafka consumer lag for topics and consumer groups');

