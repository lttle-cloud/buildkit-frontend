package api

import "encoding/json"

type AnalyzeRequestConfig struct {
	SourceDir      string
	PlanOutputDir  string
	BuildCommand   string
	StartCommand   string
	ConfigFilePath string
	Envs           []string
}

func (config *AnalyzeRequestConfig) Encode() (string, error) {
	json, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(json), nil
}

func (config *AnalyzeRequestConfig) Decode(jsonString string) error {
	return json.Unmarshal([]byte(jsonString), config)
}
