#!/bin/bash
TOPICS="emoji-reaction-topic,emoji-output-topic"

# Convert the comma-separated string into an array
IFS=',' read -r -a TOPIC_ARRAY <<< "$TOPICS"
cd ~/opt/bitnami/kafka/bin
# Check if each topic exists and create if it doesn't
for TOPIC in "${TOPIC_ARRAY[@]}"; do
  echo "Processing topic: $TOPIC"
  if ./kafka-topics.sh --list --bootstrap-server localhost:9092 | grep -q "$TOPIC"; then
    echo "Topic '$TOPIC' already exists."
  else
    echo "Creating topic '$TOPIC'..."
    ./kafka-topics.sh --bootstrap-server localhost:9092 --create --topic $TOPIC --partitions 1 --replication-factor 1
  fi
done