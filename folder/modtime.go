package folder

import (
	"os"
	"strings"
	"time"
)

const MetadataKeyModTime = "fileutil-mtime"

func CloneTime(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	value := *t
	return &value
}

func MergeMetadataWithModTime(metadata map[string]string, modTime *time.Time) map[string]string {
	if len(metadata) == 0 && modTime == nil {
		return nil
	}

	merged := make(map[string]string, len(metadata)+1)
	for key, value := range metadata {
		merged[key] = value
	}
	if modTime != nil {
		merged[MetadataKeyModTime] = modTime.UTC().Format(time.RFC3339Nano)
	}
	return merged
}

func ModTimeFromMetadata(metadata map[string]string) (*time.Time, bool) {
	for key, value := range metadata {
		if !strings.EqualFold(key, MetadataKeyModTime) {
			continue
		}

		parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(value))
		if err != nil {
			return nil, false
		}
		return &parsed, true
	}
	return nil, false
}

func ResolveModTime(explicit *time.Time, metadata map[string]string, fallback *time.Time) *time.Time {
	if explicit != nil {
		return CloneTime(explicit)
	}
	if fromMetadata, ok := ModTimeFromMetadata(metadata); ok {
		return fromMetadata
	}
	return CloneTime(fallback)
}

func ApplyLocalModTime(localPath string, modTime *time.Time) error {
	if modTime == nil {
		return nil
	}
	return os.Chtimes(localPath, *modTime, *modTime)
}
