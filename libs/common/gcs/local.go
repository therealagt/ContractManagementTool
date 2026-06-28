package gcs

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalClient stores objects on the local filesystem for dev without GCP.
type LocalClient struct {
	root   string
	bucket string
}

func NewLocalClient(root, bucket string) (*LocalClient, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &LocalClient{root: root, bucket: bucket}, nil
}

func (c *LocalClient) Close() error { return nil }

func (c *LocalClient) Upload(ctx context.Context, objectPath string, r io.Reader, _ string) error {
	_ = ctx
	path := filepath.Join(c.root, objectPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	return nil
}

func (c *LocalClient) Download(ctx context.Context, objectPath string) ([]byte, error) {
	_ = ctx
	return os.ReadFile(filepath.Join(c.root, objectPath))
}

func (c *LocalClient) FullPath(objectPath string) string {
	return fmt.Sprintf("file://%s/%s", c.bucket, objectPath)
}

func (c *LocalClient) Delete(ctx context.Context, objectPath string) error {
	_ = ctx
	return os.Remove(filepath.Join(c.root, objectPath))
}
