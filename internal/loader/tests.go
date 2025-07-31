package loader

import (
	"fmt"
	"os"

	"github.com/reegnz/policy-bot-tests/internal/models"
	"gopkg.in/yaml.v2"
)

// LoadTestFile loads and parses a test configuration file
func LoadTestFile(fileName string) (*models.TestFile, error) {
	testFile, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to load file %s: %w", fileName, err)
	}
	var tests models.TestFile
	if err := yaml.UnmarshalStrict(testFile, &tests); err != nil {
		return nil, fmt.Errorf("failed to unmarshal .policy-tests.yml: %w", err)
	}
	return &tests, nil
}
