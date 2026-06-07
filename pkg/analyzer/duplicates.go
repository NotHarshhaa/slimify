package analyzer

// DuplicateFile represents a file found in multiple layers.
type DuplicateFile struct {
	// Path is the file path within the image.
	Path string `json:"path"`
	// Size is the file size in bytes.
	Size int64 `json:"size"`
	// Layers lists the layer indices where this file appears.
	Layers []int `json:"layers"`
}

// DetectDuplicates finds files that appear in more than one layer.
// This catches the common case where files are silently copied across
// layers (e.g., in `RUN apt-get` chains or repeated COPY instructions).
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

	return duplicates
}
