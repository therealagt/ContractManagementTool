package monitoring

import (
	"context"
	"fmt"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"google.golang.org/genproto/googleapis/api/metric"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const metricPrefix = "custom.googleapis.com/contract/"

type Publisher struct {
	projectID string
	client    *monitoring.MetricClient
}

func NewPublisher(ctx context.Context, projectID string) (*Publisher, error) {
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("monitoring client: %w", err)
	}
	return &Publisher{projectID: projectID, client: client}, nil
}

func (p *Publisher) Close() error {
	return p.client.Close()
}

func (p *Publisher) PublishCount(ctx context.Context, metricName string, labels map[string]string, value int64) error {
	return p.write(ctx, metricName, labels, value)
}

func (p *Publisher) write(ctx context.Context, metricName string, labels map[string]string, value int64) error {
	now := timestamppb.Now()
	req := &monitoringpb.CreateTimeSeriesRequest{
		Name: "projects/" + p.projectID,
		TimeSeries: []*monitoringpb.TimeSeries{{
			Metric: &metric.Metric{
				Type:   metricPrefix + metricName,
				Labels: labels,
			},
			Resource: &monitoredres.MonitoredResource{
				Type: "global",
				Labels: map[string]string{
					"project_id": p.projectID,
				},
			},
			Points: []*monitoringpb.Point{{
				Interval: &monitoringpb.TimeInterval{
					EndTime: now,
				},
				Value: &monitoringpb.TypedValue{
					Value: &monitoringpb.TypedValue_Int64Value{Int64Value: value},
				},
			}},
		}},
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return p.client.CreateTimeSeries(ctx, req)
}

// NoopPublisher discards metrics (local dev).
type NoopPublisher struct{}

func (NoopPublisher) PublishCount(context.Context, string, map[string]string, int64) error {
	return nil
}

func (NoopPublisher) Close() error { return nil }
