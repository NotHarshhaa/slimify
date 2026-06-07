// Package dockerfile provides Dockerfile parsing and rewriting capabilities.
package dockerfile

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Instruction represents a single Dockerfile instruction.
type Instruction struct {
	// Command is the instruction keyword (FROM, RUN, COPY, etc.).
	Command string
	// Args is everything after the command keyword.
	Args string
	// Original is the original raw line(s) from the Dockerfile.
	Original string
	// LineNumber is the starting line number (1-indexed).
	LineNumber int
	// Stage is the build stage this instruction belongs to.
	Stage int
	// StageName is the AS alias of the current stage (if any).
	StageName string
}

// Dockerfile represents a parsed Dockerfile.
type Dockerfile struct {
	// Instructions is the ordered list of instructions.
	Instructions []Instruction
	// Stages is the number of build stages.
	Stages int
	// StageNames maps stage index to its AS alias.
	StageNames map[int]string
	// Raw is the original file content.
	Raw string
}

// Parse reads a Dockerfile from the given path and returns a parsed representation.
func Parse(path string) (*Dockerfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read Dockerfile: %w", err)
	}

	return ParseContent(string(data))
}

// ParseContent parses Dockerfile content from a string.
func ParseContent(content string) (*Dockerfile, error) {
	df := &Dockerfile{
		StageNames: make(map[int]string),
		Raw:        content,
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	currentStage := -1
	var currentLine strings.Builder
	var startLine int

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Handle line continuations
		if strings.HasSuffix(trimmed, "\\") {
			if currentLine.Len() == 0 {
				startLine = lineNum
			}
			currentLine.WriteString(strings.TrimSuffix(trimmed, "\\"))
			currentLine.WriteString(" ")
			continue
		}

		// Complete the line
		if currentLine.Len() > 0 {
			currentLine.WriteString(trimmed)
			trimmed = currentLine.String()
			currentLine.Reset()
		} else {
			startLine = lineNum
		}

		// Parse the instruction
		inst := parseInstruction(trimmed, startLine)

		// Track stages
		if inst.Command == "FROM" {
			currentStage++
			df.Stages = currentStage + 1

			// Check for AS alias
			parts := strings.Fields(inst.Args)
			for i, p := range parts {
				if strings.EqualFold(p, "AS") && i+1 < len(parts) {
					inst.StageName = parts[i+1]
					df.StageNames[currentStage] = parts[i+1]
					break
				}
			}
		}

		inst.Stage = currentStage
		inst.Original = trimmed
		df.Instructions = append(df.Instructions, inst)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning Dockerfile: %w", err)
	}

	return df, nil
}

// parseInstruction splits a Dockerfile line into command and arguments.
func parseInstruction(line string, lineNum int) Instruction {
	parts := strings.SplitN(line, " ", 2)
	cmd := strings.ToUpper(strings.TrimSpace(parts[0]))
	args := ""
	if len(parts) > 1 {
		args = strings.TrimSpace(parts[1])
	}

	return Instruction{
		Command:    cmd,
		Args:       args,
		LineNumber: lineNum,
	}
}

// GetBaseImage returns the base image from the first FROM instruction.
func (df *Dockerfile) GetBaseImage() string {
	for _, inst := range df.Instructions {
		if inst.Command == "FROM" {
			parts := strings.Fields(inst.Args)
			if len(parts) > 0 {
				return parts[0]
			}
		}
	}
	return ""
}

// GetRunInstructions returns all RUN instructions.
func (df *Dockerfile) GetRunInstructions() []Instruction {
	var runs []Instruction
	for _, inst := range df.Instructions {
		if inst.Command == "RUN" {
			runs = append(runs, inst)
		}
	}
	return runs
}

// IsMultiStage returns true if the Dockerfile has more than one FROM instruction.
func (df *Dockerfile) IsMultiStage() bool {
	return df.Stages > 1
}
