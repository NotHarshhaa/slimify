package analyzer

import (
	"archive/tar"
	"io"
)

// FileEntry represents a single file found in a layer.
type FileEntry struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	IsDir  bool   `json:"is_dir"`
	Mode   int64  `json:"mode"`
	Link   string `json:"link,omitempty"`
}

// LayerInfo holds analysis data for a single image layer.
type LayerInfo struct {
	// Index is the layer position (0 = base layer).
	Index int `json:"index"`
	// Instruction is the Dockerfile command that created this layer.
	Instruction string `json:"instruction"`
	// Size is the compressed layer size in bytes.
	Size int64 `json:"size"`
	// FileCount is the total number of files in this layer.
	FileCount int `json:"file_count"`
	// TopFiles are the largest files in this layer (above threshold).
	TopFiles []FileEntry `json:"top_files,omitempty"`
	// AllFiles is the complete file list (used internally, not serialized).
	AllFiles []FileEntry `json:"-"`
	// IsEmpty indicates if this is a metadata-only layer (e.g., ENV, LABEL).
	IsEmpty bool `json:"is_empty"`
}

// SizeMB returns the layer size in megabytes.
func (l *LayerInfo) SizeMB() float64 {
	return float64(l.Size) / (1024 * 1024)
}

// DeltaLabel returns a human-readable label for the layer delta.
func (l *LayerInfo) DeltaLabel() string {
	if l.Index == 0 {
		return "baseline"
	}
	if l.IsEmpty {
		return "metadata"
	}
	return formatSize(l.Size)
}

// readTarEntries reads all file entries from a tar stream.
func readTarEntries(r io.Reader) ([]FileEntry, error) {
	tr := tar.NewReader(r)
	var entries []FileEntry

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Skip corrupt entries
			continue
		}

		entry := FileEntry{
			Path:  hdr.Name,
			Size:  hdr.Size,
			IsDir: hdr.Typeflag == tar.TypeDir,
			Mode:  int64(hdr.Mode),
		}

		if hdr.Typeflag == tar.TypeLink || hdr.Typeflag == tar.TypeSymlink {
			entry.Link = hdr.Linkname
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// formatSize formats a byte count as a human-readable size string.
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return formatFloat(float64(bytes)/float64(GB)) + " GB"
	case bytes >= MB:
		return formatFloat(float64(bytes)/float64(MB)) + " MB"
	case bytes >= KB:
		return formatFloat(float64(bytes)/float64(KB)) + " KB"
	default:
		return formatFloat(float64(bytes)) + " B"
	}
}

// formatFloat formats a float with appropriate precision.
func formatFloat(f float64) string {
	if f >= 100 {
		return java_fmt(f, 0)
	}
	if f >= 10 {
		return java_fmt(f, 1)
	}
	return java_fmt(f, 1)
}

func java_fmt(f float64, prec int) string {
	if prec == 0 {
		return fmtInt(int64(f))
	}
	s := fmtInt(int64(f))
	frac := int64((f - float64(int64(f))) * 10)
	if frac < 0 {
		frac = -frac
	}
	return s + "." + fmtInt(frac)
}

func fmtInt(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
