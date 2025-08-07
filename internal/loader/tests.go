package loader

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/reegnz/policy-bot-tests/internal/models"
	"gopkg.in/yaml.v3"
)

// LoadTestFile loads and parses a single test configuration file.
// It is a convenience wrapper around LoadTestFiles.
func LoadTestFile(fileName string) (*models.TestFile, error) {
	return LoadTestFiles([]string{fileName})
}

// LoadTestFiles loads and parses multiple test configuration files and merges them.
func LoadTestFiles(paths []string) (*models.TestFile, error) {
	var fileList []string
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("failed to stat path %s: %w", path, err)
		}

		if info.IsDir() {
			err := filepath.WalkDir(path, func(s string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if strings.HasSuffix(d.Name(), ".policy-tests.yml") || strings.HasSuffix(d.Name(), ".policy-tests.yaml") {
					fileList = append(fileList, s)
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("failed to walk directory %s: %w", path, err)
			}
		} else {
			fileList = append(fileList, path)
		}
	}

	mergedTests := &models.TestFile{}
	for _, file := range fileList {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to load file %s: %w", file, err)
		}

		var node yaml.Node
		if err := yaml.Unmarshal(content, &node); err != nil {
			return nil, fmt.Errorf("failed to parse YAML in %s: %w", file, err)
		}

		var tests models.TestFile
		decoder := yaml.NewDecoder(bytes.NewReader(content))
		decoder.KnownFields(true)
		if err := decoder.Decode(&tests); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s: %w", file, err)
		}

		extractLineNumbers(&node, &tests)

		// Set filename for all test cases from this file
		for i := range tests.TestCases {
			// Get relative path from current working directory
			relPath, err := filepath.Rel(".", file)
			if err != nil {
				// Fallback to base filename if relative path fails
				tests.TestCases[i].FileName = filepath.Base(file)
			} else {
				tests.TestCases[i].FileName = relPath
			}
		}

		// Simple merge: append test cases, last defaultContext wins.
		mergedTests.TestCases = append(mergedTests.TestCases, tests.TestCases...)
		if tests.DefaultContext.Owner != "" {
			mergedTests.DefaultContext = tests.DefaultContext
		}
	}

	return mergedTests, nil
}

// extractLineNumbers extracts line numbers from YAML nodes and sets them on test cases
func extractLineNumbers(node *yaml.Node, tests *models.TestFile) {
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}

	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]

			if key.Value == "test_cases" && value.Kind == yaml.SequenceNode {
				for j, testNode := range value.Content {
					if j < len(tests.TestCases) {
						tests.TestCases[j].LineNumber = testNode.Line
					}
				}
			}
		}
	}
}
