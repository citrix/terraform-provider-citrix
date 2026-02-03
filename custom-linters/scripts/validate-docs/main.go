// Copyright © 2026. Citrix Systems, Inc.

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := validateDocs(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	_, _ = fmt.Fprintf(os.Stdout, "✓ All resources and data sources have valid subcategory metadata\n") //nolint:errcheck // stdout write error not actionable
}

func validateDocs() error {
	resourcesDir := "docs/resources"
	if err := validateDirectory(resourcesDir, "resource"); err != nil {
		return err
	}

	dataSourcesDir := "docs/data-sources"
	if err := validateDirectory(dataSourcesDir, "data source"); err != nil {
		return err
	}

	return nil
}

func validateDirectory(dir string, docType string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read %s directory: %w", dir, err)
	}

	violations := []string{}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		if err := validateDocFile(filePath, &violations); err != nil {
			return err
		}
	}

	if len(violations) > 0 {
		fmt.Fprintf(os.Stderr, "\nFound %d %s(s) without subcategory:\n", len(violations), docType)
		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "  - %s\n", v)
		}
		return fmt.Errorf("%d %s(s) are missing subcategory metadata", len(violations), docType)
	}

	return nil
}

func validateDocFile(filePath string, violations *[]string) error {
	// #nosec G304 -- filePath is constructed from internal logic, not user input
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer file.Close() //nolint:errcheck // Error not actionable in defer

	scanner := bufio.NewScanner(file)
	inFrontmatter := false
	hasSubcategory := false
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "---" {
			if lineNum == 1 {
				inFrontmatter = true
				continue
			} else if inFrontmatter {
				break
			}
		}

		if inFrontmatter && strings.HasPrefix(line, "subcategory:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				subcategory := strings.TrimSpace(strings.Trim(parts[1], `"`))
				if subcategory != "" {
					hasSubcategory = true
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	if !hasSubcategory {
		*violations = append(*violations, filePath)
	}

	return nil
}
