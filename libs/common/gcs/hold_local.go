package gcs

import (
	"context"
	"os"
	"path/filepath"
)

func (c *LocalClient) SetEventHold(ctx context.Context, objectPath string, hold bool) error {
	_ = ctx
	marker := filepath.Join(c.root, objectPath+".hold")
	if hold {
		return os.WriteFile(marker, []byte("1"), 0o644)
	}
	if err := os.Remove(marker); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (c *LocalClient) HasHold(objectPath string) bool {
	_, err := os.Stat(filepath.Join(c.root, objectPath+".hold"))
	return err == nil
}
