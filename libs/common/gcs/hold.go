package gcs

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
)

func (c *Client) SetEventHold(ctx context.Context, objectPath string, hold bool) error {
	obj := c.client.Bucket(c.bucket).Object(objectPath)
	_, err := obj.Update(ctx, storage.ObjectAttrsToUpdate{
		EventBasedHold: &hold,
	})
	if err != nil {
		return fmt.Errorf("gcs set hold: %w", err)
	}
	return nil
}
