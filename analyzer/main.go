package main

import (
	"encoding/json"
	"os"
	"path"

	"github.com/lttle-cloud/buildkit-frontend/api"

	"github.com/charmbracelet/log"

	rl "github.com/railwayapp/railpack/core"
	rlApp "github.com/railwayapp/railpack/core/app"
)

func main() {
	configString := os.Getenv("ANALYZER_REQUEST_CONFIG")
	if configString == "" {
		log.Error("error: analyze request config is required")
		os.Exit(1)
	}

	var config api.AnalyzeRequestConfig
	if err := config.Decode(configString); err != nil {
		log.Error("error: %+v\n", err)
		os.Exit(1)
	}

	if err := AnalyzeRailpack(config); err != nil {
		log.Error("error: %+v\n", err)
		os.Exit(1)
	}
}

func AnalyzeRailpack(config api.AnalyzeRequestConfig) error {
	app, err := rlApp.NewApp(config.SourceDir)
	if err != nil {
		return err
	}

	env, err := rlApp.FromEnvs(config.Envs)
	if err != nil {
		return err
	}

	generateOptions := &rl.GenerateBuildPlanOptions{
		RailpackVersion:          "dev",
		BuildCommand:             config.BuildCommand,
		StartCommand:             config.StartCommand,
		PreviousVersions:         map[string]string{},
		ConfigFilePath:           config.ConfigFilePath,
		ErrorMissingStartCommand: true,
	}

	buildResult := rl.GenerateBuildPlan(app, env, generateOptions)
	if buildResult.Success && buildResult.Plan != nil {
		railpackPlanOutputPath := path.Join(config.PlanOutputDir, "railpack-plan.json")
		railpackJson, err := json.Marshal(buildResult.Plan)
		if err == nil {
			os.WriteFile(railpackPlanOutputPath, railpackJson, 0644)
		}
	}

	buildResult.Plan = nil

	railpackInfoOutputPath := path.Join(config.PlanOutputDir, "railpack-info.json")
	infoJson, err := json.Marshal(buildResult)
	if err == nil {
		os.WriteFile(railpackInfoOutputPath, infoJson, 0644)
	}

	return nil
}
