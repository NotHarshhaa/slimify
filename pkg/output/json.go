package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/NotHarshhaa/slimify/pkg/analyzer"
)

// PrintAuditJSON outputs the audit report as JSON.
func PrintAuditJSON(report *analyzer.AuditReport) error {
	return printJSON(report)
}

// PrintCompareJSON outputs the compare report as JSON.
func PrintCompareJSON(report *analyzer.CompareReport) error {
	return printJSON(report)
}

// printJSON serializes any value as indented JSON to stdout.
func printJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}

// ToJSON serializes any value as indented JSON to a string.
func ToJSON(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}
