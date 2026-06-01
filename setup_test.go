package heijitu_test

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// TestYAMLDependencyAvailable は gopkg.in/yaml.v3 がモジュールに追加されており、
// インポートして正常に使用できることを確認するスモークテストである。
// このテストがビルドに成功すること自体が、go.mod / go.sum の正しい生成を示す。
func TestYAMLDependencyAvailable(t *testing.T) {
	input := `
key: value
number: 42
`
	var result map[string]any
	err := yaml.Unmarshal([]byte(input), &result)

	if err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("key: expected %q, got %v", "value", result["key"])
	}
	if result["number"] != 42 {
		t.Errorf("number: expected 42, got %v", result["number"])
	}
}

// TestYAMLDependencyUnmarshalError は yaml.v3 が不正な YAML に対してエラーを返すことを確認する。
func TestYAMLDependencyUnmarshalError(t *testing.T) {
	var out any
	err := yaml.Unmarshal([]byte("key: [unclosed"), &out)

	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}
