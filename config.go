package heijitu

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// config は設定ファイルから読み込む設定を保持する。
type config struct {
	ExcludedDates []MonthDay `yaml:"excluded_dates" json:"excluded_dates"`
}

const (
	extYAML = ".yaml"
	extYML  = ".yml"
	extJSON = ".json"
)

// loadConfig は指定されたパスの設定ファイルを読み込み config を返す。
// サポートする拡張子: .yaml, .yml, .json
func loadConfig(path string) (*config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(path)
	var cfg config
	switch ext {
	case extYAML, extYML:
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	case extJSON:
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported config file extension: %q", ext)
	}
	return &cfg, nil
}
