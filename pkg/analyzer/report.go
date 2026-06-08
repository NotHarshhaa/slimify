package analyzer

import (
	"github.com/NotHarshhaa/slimify/pkg/ecosystem"
)

// AuditReport is the complete result of an image audit.
type AuditReport struct {
	// ImageRef is the image reference that was audited.
	ImageRef string `json:"image_ref"`
	// TotalSize is the total compressed image size in bytes.
	TotalSize int64 `json:"total_size"`
	// TotalSizeMB is the total size in megabytes.
	TotalSizeMB float64 `json:"total_size_mb"`
	// LayerCount is the total number of non-empty layers.
	LayerCount int `json:"layer_count"`
	// Layers contains per-layer analysis.
	Layers []LayerInfo `json:"layers"`
	// Ecosystems detected in the image.
	Ecosystems *ecosystem.DetectResult `json:"ecosystems"`
	// Duplicates are files found in multiple layers.
	Duplicates []DuplicateFile `json:"duplicates,omitempty"`
	// SecretFiles lists files that look like they may contain secrets.
	SecretFiles []string `json:"secret_files,omitempty"`
	// Recommendations are actionable suggestions to reduce image size.
	Recommendations []Recommendation `json:"recommendations"`
	// SavingsMB is the total estimated savings in megabytes.
	SavingsMB float64 `json:"savings_mb"`
	// SavingsPercent is the percentage of total size that could be saved.
	SavingsPercent float64 `json:"savings_percent"`
}

// Recommendation is a single actionable suggestion.
type Recommendation struct {
	// Title is a short one-line summary.
	Title string `json:"title"`
	// Detail is a longer explanation.
	Detail string `json:"detail"`
	// SavingsMB is the estimated savings from applying this recommendation.
	SavingsMB float64 `json:"savings_mb"`
	// Priority is the recommendation priority (1 = highest).
	Priority int `json:"priority"`
}

// CompareReport holds the result of comparing two images.
type CompareReport struct {
	// ImageA is the first image reference.
	ImageA string `json:"image_a"`
	// ImageB is the second image reference.
	ImageB string `json:"image_b"`
	// SizeA is the size of image A in bytes.
	SizeA int64 `json:"size_a"`
	// SizeB is the size of image B in bytes.
	SizeB int64 `json:"size_b"`
	// Reduction is the size difference in bytes (positive = B is smaller).
	Reduction int64 `json:"reduction"`
	// ReductionPercent is the percentage reduction.
	ReductionPercent float64 `json:"reduction_percent"`
	// LayersA is the number of layers in image A.
	LayersA int `json:"layers_a"`
	// LayersB is the number of layers in image B.
	LayersB int `json:"layers_b"`
	// NewLayersInB is the count of layers in B not present in A.
	NewLayersInB int `json:"new_layers_in_b"`
	// RemovedLayersInB is the count of layers in A not present in B.
	RemovedLayersInB int `json:"removed_layers_in_b"`
	// SharedBaseLayers is the count of layers shared between both images.
	SharedBaseLayers int `json:"shared_base_layers"`
}

// CompareImages compares two Docker images and returns a comparison report.
func (a *ImageAnalyzer) CompareImages(imageRefA, imageRefB string, isRemote bool) (*CompareReport, error) {
	imgA, err := a.loadImage(imageRefA, isRemote)
	if err != nil {
		return nil, err
	}
	imgB, err := a.loadImage(imageRefB, isRemote)
	if err != nil {
		return nil, err
	}

	manifestA, err := imgA.Manifest()
	if err != nil {
		return nil, err
	}
	manifestB, err := imgB.Manifest()
	if err != nil {
		return nil, err
	}

	var sizeA, sizeB int64
	for _, l := range manifestA.Layers {
		sizeA += l.Size
	}
	for _, l := range manifestB.Layers {
		sizeB += l.Size
	}

	layersA, _ := imgA.Layers()
	layersB, _ := imgB.Layers()

	// Find shared layers by digest
	digestsA := make(map[string]bool)
	for _, l := range layersA {
		d, err := l.Digest()
		if err == nil {
			digestsA[d.String()] = true
		}
	}

	shared := 0
	for _, l := range layersB {
		d, err := l.Digest()
		if err == nil {
			if digestsA[d.String()] {
				shared++
			}
		}
	}

	reduction := sizeA - sizeB
	var reductionPct float64
	if sizeA > 0 {
		reductionPct = float64(reduction) / float64(sizeA) * 100
	}

	return &CompareReport{
		ImageA:           imageRefA,
		ImageB:           imageRefB,
		SizeA:            sizeA,
		SizeB:            sizeB,
		Reduction:        reduction,
		ReductionPercent: reductionPct,
		LayersA:          len(layersA),
		LayersB:          len(layersB),
		NewLayersInB:     len(layersB) - shared,
		RemovedLayersInB: len(layersA) - shared,
		SharedBaseLayers: shared,
	}, nil
}
