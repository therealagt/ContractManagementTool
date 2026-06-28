package gcs

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

type Client struct {
	bucket string
	client *storage.Client
}

func NewClient(ctx context.Context, bucket string) (*Client, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcs client: %w", err)
	}
	return &Client{bucket: bucket, client: client}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func StagingObjectPath(contractID string) string {
	return fmt.Sprintf("staging/%s.pdf", contractID)
}

func ArchiveObjectPath(contractID string) string {
	return fmt.Sprintf("archive/%s.pdf", contractID)
}

func (c *Client) Upload(ctx context.Context, objectPath string, r io.Reader, contentType string) error {
	w := c.client.Bucket(c.bucket).Object(objectPath).NewWriter(ctx)
	w.ContentType = contentType
	if _, err := io.Copy(w, r); err != nil {
		_ = w.Close()
		return fmt.Errorf("gcs upload: %w", err)
	}
	return w.Close()
}

func (c *Client) Download(ctx context.Context, objectPath string) ([]byte, error) {
	r, err := c.client.Bucket(c.bucket).Object(objectPath).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcs download: %w", err)
	}
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("gcs read: %w", err)
	}
	return data, nil
}

func (c *Client) FullPath(objectPath string) string {
	return fmt.Sprintf("gs://%s/%s", c.bucket, objectPath)
}

func (c *Client) Delete(ctx context.Context, objectPath string) error {
	if err := c.client.Bucket(c.bucket).Object(objectPath).Delete(ctx); err != nil {
		return fmt.Errorf("gcs delete: %w", err)
	}
	return nil
}
