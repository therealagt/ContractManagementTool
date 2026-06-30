package gcs

import (
	"fmt"
	"strings"
)

// StagingObjectPathFromFullPath returns the staging object key (e.g. staging/<id>.pdf).
func StagingObjectPathFromFullPath(fullPath string) (string, error) {
	_, objectPath, err := ParseFullPath(fullPath)
	if err != nil {
		return "", err
	}
	const prefix = "staging/"
	if !strings.HasPrefix(objectPath, prefix) {
		return "", fmt.Errorf("not a staging object path: %s", fullPath)
	}
	return objectPath, nil
}

// ParseFullPath splits gs://bucket/object into bucket and object path.
func ParseFullPath(fullPath string) (bucket, objectPath string, err error) {
	const prefix = "gs://"
	if !strings.HasPrefix(fullPath, prefix) {
		return "", "", fmt.Errorf("invalid gcs path: %s", fullPath)
	}
	rest := strings.TrimPrefix(fullPath, prefix)
	slash := strings.Index(rest, "/")
	if slash < 0 {
		return "", "", fmt.Errorf("invalid gcs path: %s", fullPath)
	}
	return rest[:slash], rest[slash+1:], nil
}
