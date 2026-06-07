// Package analyzer provides Docker image analysis capabilities.
// It reads OCI-compatible images (local or remote), extracts layer
// information, builds file trees, and produces audit reports.
package analyzer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"

	"github.com/NotHarshhaa/slimify/pkg/ecosystem"
)

// ImageAnalyzer loads and analyzes Docker images.
type ImageAnalyzer struct {
	// TopFilesPerLayer controls how many top files to show per layer.
	TopFilesPerLayer int
	// ThresholdBytes is the minimum file size to flag.
	ThresholdBytes int64
}

// NewImageAnalyzer creates an analyzer with the given settings.
func NewImageAnalyzer(topFiles int, thresholdMB float64) *ImageAnalyzer {
	return &ImageAnalyzer{
		TopFilesPerLayer: topFiles,
		ThresholdBytes:   int64(thresholdMB * 1024 * 1024),
	}
}

// AnalyzeImage loads and analyzes a Docker image, returning a full report.
func (a *ImageAnalyzer) AnalyzeImage(imageRef string, isRemote bool) (*AuditReport, error) {
	img, err := a.loadImage(imageRef, isRemote)
	if err != nil {
		return nil, fmt.Errorf("failed to load image %q: %w", imageRef, err)
	}

	return a.analyze(imageRef, img)
}

// loadImage loads a Docker image from the local daemon or a remote registry.
func (a *ImageAnalyzer) loadImage(imageRef string, isRemote bool) (v1.Image, error) {
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return nil, fmt.Errorf("invalid image reference: %w", err)
	}

	if isRemote {
		return remote.Image(ref)
	}

	// Try local daemon first
	img, err := daemon.Image(ref)
	if err == nil {
		return img, nil
	}

	// Try as a local tarball
	img, err = tarball.ImageFromPath(imageRef, nil)
	if err == nil {
		return img, nil
	}

	// Try pulling via crane as fallback
	img, err = crane.Pull(imageRef)
	if err != nil {
		return nil, fmt.Errorf("could not load image from daemon, tarball, or registry: %w", err)
	}
	return img, nil
}

// analyze performs the actual image analysis.
func (a *ImageAnalyzer) analyze(imageRef string, img v1.Image) (*AuditReport, error) {
	manifest, err := img.Manifest()
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	configFile, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("failed to read layers: %w", err)
	}

	// Calculate total image size
	var totalSize int64
	for _, l := range manifest.Layers {
		totalSize += l.Size
	}

	// Analyze each layer
	allFiles := make(map[string]bool)
	var layerInfos []LayerInfo

	for i, layer := range layers {
		instruction := "unknown"
		if i < len(configFile.History) {
			instruction = cleanInstruction(configFile.History[i].CreatedBy)
		}

		size, err := layer.Size()
		if err != nil {
			size = 0
		}

		files, err := extractLayerFiles(layer)
		if err != nil {
			files = []FileEntry{}
		}

		// Track all files for ecosystem detection
		for _, f := range files {
			allFiles[f.Path] = true
		}

		// Sort files by size descending
		sort.Slice(files, func(a, b int) bool {
			return files[a].Size > files[b].Size
		})

		// Get top N files
		topFiles := files
		if len(topFiles) > a.TopFilesPerLayer {
			topFiles = topFiles[:a.TopFilesPerLayer]
		}

		// Filter by threshold
		var flaggedFiles []FileEntry
		for _, f := range topFiles {
			if f.Size >= a.ThresholdBytes {
				flaggedFiles = append(flaggedFiles, f)
			}
		}

		layerInfo := LayerInfo{
			Index:        i,
			Instruction:  instruction,
			Size:         size,
			FileCount:    len(files),
			TopFiles:     flaggedFiles,
			AllFiles:     files,
			IsEmpty:      configFile.History[i].EmptyLayer,
		}

		layerInfos = append(layerInfos, layerInfo)
	}

	// Detect ecosystems from all files in the image
	var filePaths []string
	for p := range allFiles {
		filePaths = append(filePaths, p)
	}
	ecosystems := ecosystem.DetectFromFiles(filePaths)

	// Detect duplicate files across layers
	duplicates := DetectDuplicates(layerInfos)

	// Generate recommendations
	bloatPatterns := ecosystem.GetBloatPatterns(ecosystems)
	recommendations := generateRecommendations(layerInfos, ecosystems, bloatPatterns, duplicates)

	// Calculate potential savings
	var savingsMB float64
	for _, rec := range recommendations {
		savingsMB += rec.SavingsMB
	}

	report := &AuditReport{
		ImageRef:       imageRef,
		TotalSize:      totalSize,
		Layers:         layerInfos,
		Ecosystems:     ecosystems,
		Duplicates:     duplicates,
		Recommendations: recommendations,
		SavingsMB:      savingsMB,
		SavingsPercent: 0,
	}

	if totalSize > 0 {
		report.SavingsPercent = (savingsMB / float64(totalSize) * 1024 * 1024) * 100
		// Cap at 100%
		if report.SavingsPercent > 100 {
			report.SavingsPercent = float64(int(savingsMB / (float64(totalSize) / (1024 * 1024)) * 100))
		}
	}

	return report, nil
}

// extractLayerFiles reads all file entries from a layer's tarball.
func extractLayerFiles(layer v1.Layer) ([]FileEntry, error) {
	rc, err := layer.Uncompressed()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return readTarEntries(rc)
}

// cleanInstruction cleans up a Docker history instruction for display.
func cleanInstruction(instruction string) string {
	// Remove the /bin/sh -c prefix that Docker adds
	instruction = strings.TrimPrefix(instruction, "/bin/sh -c ")
	instruction = strings.TrimPrefix(instruction, "#(nop) ")
	instruction = strings.TrimSpace(instruction)

	// Truncate very long instructions
	if len(instruction) > 60 {
		instruction = instruction[:57] + "..."
	}

	return instruction
}

// generateRecommendations produces actionable recommendations from the analysis.
func generateRecommendations(layers []LayerInfo, eco *ecosystem.DetectResult, patterns []ecosystem.BloatPattern, duplicates []DuplicateFile) []Recommendation {
	var recs []Recommendation

	// Check for bloat in COPY layers
	for _, layer := range layers {
		if !strings.Contains(strings.ToUpper(layer.Instruction), "COPY") {
			continue
		}

		var copyBloatMB float64
		for _, f := range layer.AllFiles {
			if matched, pattern := ecosystem.MatchesBloatPattern(f.Path, patterns); matched {
				copyBloatMB += float64(f.Size) / (1024 * 1024)
				_ = pattern
			}
		}

		if copyBloatMB > 1 {
			recs = append(recs, Recommendation{
				Title:     "Generate .dockerignore (run slimify fix)",
				Detail:    fmt.Sprintf("Remove bloat from COPY context in layer: %s", layer.Instruction),
				SavingsMB: copyBloatMB,
				Priority:  2,
			})
		}
	}

	// Check for multi-stage build opportunity
	if eco.HasEcosystem(ecosystem.NodeJS) {
		var nodeModulesSize float64
		for _, layer := range layers {
			for _, f := range layer.AllFiles {
				if strings.HasPrefix(f.Path, "node_modules/") || strings.Contains(f.Path, "/node_modules/") {
					nodeModulesSize += float64(f.Size) / (1024 * 1024)
				}
			}
		}
		if nodeModulesSize > 10 {
			recs = append(recs, Recommendation{
				Title:     "Switch to multi-stage build",
				Detail:    "Discard build-stage node_modules and only copy production artifacts",
				SavingsMB: nodeModulesSize * 0.6, // estimate 60% savings
				Priority:  1,
			})
		}
	}

	// Check for alpine base opportunity
	for _, layer := range layers {
		if layer.Index == 0 && layer.Size > 100*1024*1024 {
			instruction := strings.ToLower(layer.Instruction)
			if strings.Contains(instruction, "from") && !strings.Contains(instruction, "alpine") && !strings.Contains(instruction, "slim") && !strings.Contains(instruction, "distroless") {
				estimatedSavings := float64(layer.Size) / (1024 * 1024) * 0.35
				recs = append(recs, Recommendation{
					Title:     "Use an alpine or slim base image",
					Detail:    fmt.Sprintf("Current base layer is %.0f MB — switch to alpine/slim variant", float64(layer.Size)/(1024*1024)),
					SavingsMB: estimatedSavings,
					Priority:  1,
				})
			}
		}
	}

	// Check for RUN layer consolidation
	var consecutiveRuns int
	var runLayerSize float64
	for _, layer := range layers {
		if strings.HasPrefix(strings.ToUpper(layer.Instruction), "RUN") ||
			strings.Contains(layer.Instruction, "apt-get") ||
			strings.Contains(layer.Instruction, "apk add") ||
			strings.Contains(layer.Instruction, "yum install") {
			consecutiveRuns++
			runLayerSize += float64(layer.Size) / (1024 * 1024)
		}
	}
	if consecutiveRuns > 2 {
		recs = append(recs, Recommendation{
			Title:     "Merge RUN instructions + cleanup in one layer",
			Detail:    fmt.Sprintf("Found %d separate RUN layers — merge to eliminate dead layer space", consecutiveRuns),
			SavingsMB: runLayerSize * 0.15, // estimate 15% savings from merging
			Priority:  3,
		})
	}

	// Check duplicates
	if len(duplicates) > 0 {
		var dupSize float64
		for _, d := range duplicates {
			dupSize += float64(d.Size) / (1024 * 1024)
		}
		if dupSize > 1 {
			recs = append(recs, Recommendation{
				Title:     "Consolidate duplicate files across layers",
				Detail:    fmt.Sprintf("Found %d files duplicated across layers", len(duplicates)),
				SavingsMB: dupSize,
				Priority:  4,
			})
		}
	}

	// Sort by priority
	sort.Slice(recs, func(i, j int) bool {
		return recs[i].Priority < recs[j].Priority
	})

	return recs
}
