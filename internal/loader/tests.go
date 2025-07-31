package loader

import (
	"bytes"
	"fmt"
	"os"

	"github.com/reegnz/policy-bot-tests/internal/models"
	"gopkg.in/yaml.v3"
)

// LoadTestFile loads and parses a test configuration file
func LoadTestFile(fileName string) (*models.TestFile, error) {
	testFile, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to load file %s: %w", fileName, err)
	}

	var node yaml.Node
	if err := yaml.Unmarshal(testFile, &node); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	var tests models.TestFile
	decoder := yaml.NewDecoder(bytes.NewReader(testFile))
	decoder.KnownFields(true)
	if err := decoder.Decode(&tests); err != nil {
		return nil, fmt.Errorf("failed to unmarshal .policy-tests.yml: %w", err)
	}

	// Extract line numbers for test cases
	extractLineNumbers(&node, &tests)

	return &tests, nil
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

			if key.Value == "testCases" && value.Kind == yaml.SequenceNode {
				for j, testNode := range value.Content {
					if j < len(tests.TestCases) {
						tests.TestCases[j].LineNumber = testNode.Line
					}
				}
			}
		}
	}
}
