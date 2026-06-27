package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	gcpubsub "cloud.google.com/go/pubsub/v2"
)

type ExtractionRequested struct {
	ContractID    string `json:"contract_id"`
	Type          string `json:"type"`
	GCSPath       string `json:"gcs_path"`
	SchemaVersion string `json:"schema_version"`
}

type Publisher struct {
	client *gcpubsub.Client
	topic  string
}

func NewPublisher(ctx context.Context, projectID, topic string) (*Publisher, error) {
	client, err := gcpubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("pubsub client: %w", err)
	}
	return &Publisher{client: client, topic: topic}, nil
}

func (p *Publisher) Close() error {
	return p.client.Close()
}

func (p *Publisher) PublishExtraction(ctx context.Context, msg ExtractionRequested) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal extraction message: %w", err)
	}
	publisher := p.client.Publisher(p.topic)
	result := publisher.Publish(ctx, &gcpubsub.Message{Data: data})
	_, err = result.Get(ctx)
	if err != nil {
		return fmt.Errorf("publish extraction: %w", err)
	}
	return nil
}
