package analyzer

import "sort"

// DuplicateFile represents a file found in multiple layers.
type DuplicateFile struct {
	// Path is the file path within the image.
	Path string `json:"path"`
	// Size is the file size in bytes.
	Size int64 `json:"size"`
	// Layers lists the layer indices where this file appears.
	Layers []int `json:"layers"`
}

// maxDuplicates caps the number of reported duplicates to avoid flooding output.
const maxDuplicates = 50

// DetectDuplicates finds files that appear in more than one layer.
// This catches the common case where files are silently copied across
// layers (e.g., in `RUN apt-get` chains or repeated COPY instructions).
// Results are sorted by size descending so the biggest wasted space appears first.
func DetectDuplicates(layers []LayerInfo) []DuplicateFile {
	// Map: file path -> list of layer indices and size
	type fileOccurrence struct {
		layers []int
		size   int64
	}

	fileMap := make(map[string]*fileOccurrence)

	for _, layer := range layers {
		// Track which files we've already seen in this layer
		seenInLayer := make(map[string]bool)

		for _, f := range layer.AllFiles {
			if f.IsDir {
				continue
			}
			if seenInLayer[f.Path] {
				continue
			}
			seenInLayer[f.Path] = true

			if occ, ok := fileMap[f.Path]; ok {
				occ.layers = append(occ.layers, layer.Index)
				// Use the latest size
				occ.size = f.Size
			} else {
				fileMap[f.Path] = &fileOccurrence{
					layers: []int{layer.Index},
					size:   f.Size,
				}
			}
		}
	}

	// Collect files that appear in more than one layer
	var duplicates []DuplicateFile
	for path, occ := range fileMap {
		if len(occ.layers) > 1 && occ.size > 0 {
			duplicates = append(duplicates, DuplicateFile{
				Path:   path,
				Size:   occ.size,
				Layers: occ.layers,
			})
		}
	}

	// Sort by size descending — biggest waste first.
	sort.Slice(duplicates, func(i, j int) bool {
		return duplicates[i].Size > duplicates[j].Size
	})

	// Cap results to avoid flooding output.
	if len(duplicates) > maxDuplicates {
		duplicates = duplicates[:maxDuplicates]
	}

	return duplicates
}
