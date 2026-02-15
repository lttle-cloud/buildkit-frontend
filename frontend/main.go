package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/lttle-cloud/buildkit-frontend/api"

	"github.com/charmbracelet/log"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/frontend/gateway/client"
	gw "github.com/moby/buildkit/frontend/gateway/grpcclient"
	"github.com/moby/buildkit/solver/pb"
	"github.com/moby/buildkit/util/appcontext"
)

const (
	DEFAULT_RAILPACK_FRONTEND_IMAGE = "europe-docker.pkg.dev/azin-dev/builder/railpack-frontend:latest"
	DEFAULT_ANALYZER_IMAGE          = "europe-docker.pkg.dev/azin-dev/builder/buildkit-analyzer:latest"
)

const (
	analyzerImageKey         = "analyzer-image"
	railpackFrontendImageKey = "railpack-frontend-image"

	contextKey   = "context"
	gitRefKey    = "git-ref"
	gitSubdirKey = "git-subdir"

	// Report back options
	reportBuildIdKey   = "report-build-id"
	reportBaseUrlKey   = "report-base-url"
	reportAuthTokenKey = "report-auth-token"

	// Passtrough options to the analyzer
	startCommandKey   = "start-command"
	buildCommandKey   = "build-command"
	configFilePathKey = "config-file-path"
	envNamesKey       = "env-names"

	// Passtrough options to the railpack frontend
	secretsHashKey = "secrets-hash"
	cacheKeyKey    = "cache-key"
	githubTokenKey = "github-token"
)

const (
	railpackSecretsHash = "secrets-hash"
	railpackCacheKey    = "cache-key"
	railpackGithubToken = "github-token"
)

func main() {
	StartFrontend()
}

func StartFrontend() {
	ctx := appcontext.Context()
	if err := gw.RunFromEnvironment(ctx, Build); err != nil {
		log.Error("error: %+v\n", err)
		os.Exit(1)
	}
}

func Build(ctx context.Context, c client.Client) (*client.Result, error) {
	opts := c.BuildOpts().Opts

	gitRepoUrl := opts[contextKey]
	if gitRepoUrl == "" {
		return nil, fmt.Errorf("context is required")
	}

	gitRef := ""
	if gitRefOpt, ok := opts[gitRefKey]; ok {
		gitRef = gitRefOpt
	}

	gitSubdir := ""
	if gitSubdirOpt, ok := opts[gitSubdirKey]; ok {
		gitSubdir = gitSubdirOpt
	}

	secretsHash := ""
	if secretsHashOpt, ok := opts[secretsHashKey]; ok {
		secretsHash = secretsHashOpt
	}

	cacheKey := ""
	if cacheKeyOpt, ok := opts[cacheKeyKey]; ok {
		cacheKey = cacheKeyOpt
	}

	githubToken := ""
	if githubTokenOpt, ok := opts[githubTokenKey]; ok {
		githubToken = githubTokenOpt
	}

	startCommand := ""
	if startCommandOpt, ok := opts[startCommandKey]; ok {
		startCommand = startCommandOpt
	}

	buildCommand := ""
	if buildCommandOpt, ok := opts[buildCommandKey]; ok {
		buildCommand = buildCommandOpt
	}

	configFilePath := ""
	if configFilePathOpt, ok := opts[configFilePathKey]; ok {
		configFilePath = configFilePathOpt
	}

	reportBuildId := ""
	if reportBuildIdOpt, ok := opts[reportBuildIdKey]; ok {
		reportBuildId = reportBuildIdOpt
	}

	reportBaseUrl := ""
	if reportBaseUrlOpt, ok := opts[reportBaseUrlKey]; ok {
		reportBaseUrl = reportBaseUrlOpt
	}

	reportAuthToken := ""
	if reportAuthTokenOpt, ok := opts[reportAuthTokenKey]; ok {
		reportAuthToken = reportAuthTokenOpt
	}

	envNames := []string{}
	if envNamesOpt, ok := opts[envNamesKey]; ok {
		for _, envName := range strings.Split(envNamesOpt, ",") {
			envName = strings.TrimSpace(envName)
			if envName == "" {
				continue
			}

			envNames = append(envNames, envName)
		}
	}

	analyzerImage := DEFAULT_ANALYZER_IMAGE
	if analyzerImageOpt, ok := opts[analyzerImageKey]; ok {
		analyzerImage = analyzerImageOpt
	}

	railpackFrontendImage := DEFAULT_RAILPACK_FRONTEND_IMAGE
	if railpackFrontendImageOpt, ok := opts[railpackFrontendImageKey]; ok {
		railpackFrontendImage = railpackFrontendImageOpt
	}

	git := llb.Git(gitRepoUrl, gitRef, llb.KeepGitDir(), llb.GitSubDir(gitSubdir))

	analyzerRequestConfig := api.AnalyzeRequestConfig{
		SourceDir:      "/src",
		PlanOutputDir:  "/out",
		BuildCommand:   buildCommand,
		StartCommand:   startCommand,
		ConfigFilePath: configFilePath,
		Envs:           envNames,
	}

	analyzerRequestConfigString, err := analyzerRequestConfig.Encode()
	if err != nil {
		return nil, fmt.Errorf("error encoding analyzer request config: %+v", err)
	}

	analyzerRunOptions := []llb.RunOption{
		llb.Args([]string{"/usr/bin/analyzer"}),
		llb.AddEnv("ANALYZER_REQUEST_CONFIG", analyzerRequestConfigString),
		llb.Dir("/src"),
		llb.AddMount("/src", git, llb.Readonly),
		llb.AddMount("/out", llb.Scratch()),
		llb.AddMount("/tmp", llb.Scratch(), llb.Tmpfs()),
		llb.AddEnv("TMPDIR", "/tmp"),
	}

	for _, envName := range envNames {
		analyzerRunOptions = append(analyzerRunOptions, llb.AddSecret(envName, llb.SecretAsEnv(true), llb.SecretAsEnvName(envName)))
	}

	analyzer := llb.Image(analyzerImage, llb.LinuxAmd64)
	run := analyzer.Run(analyzerRunOptions...)
	out := run.AddMount("/out", llb.Scratch())

	ctxDef, _ := git.Marshal(ctx)
	dfDef, _ := out.Marshal(ctx)

	reportClient := api.NewReportClient(reportBaseUrl, reportAuthToken)

	if reportClient.IsConfigured() && reportBuildId != "" {
		res, _ := c.Solve(ctx, client.SolveRequest{Definition: dfDef.ToPB()})
		ref, _ := res.SingleRef()

		buildInfo, _ := ref.ReadFile(ctx, client.ReadRequest{
			Filename: "railpack-info.json",
		})
		buildInfoString := string(buildInfo)

		buildPlan, _ := ref.ReadFile(ctx, client.ReadRequest{
			Filename: "railpack-plan.json",
		})
		buildPlanString := string(buildPlan)

		reportClient.ReportAsync(&api.ReportRequest{
			BuildId:             reportBuildId,
			SerializedBuildInfo: buildInfoString,
			SerializedPlan:      buildPlanString,
		})
	}

	res, err := c.Solve(ctx, client.SolveRequest{
		Frontend: "gateway.v0",
		FrontendOpt: map[string]string{
			"source":            railpackFrontendImage,
			railpackGithubToken: githubToken,
			railpackCacheKey:    cacheKey,
			railpackSecretsHash: secretsHash,
		},
		FrontendInputs: map[string]*pb.Definition{
			"context":    ctxDef.ToPB(),
			"dockerfile": dfDef.ToPB(),
		}})

	reportClient.WaitForAllAsyncReports()

	return res, err
}
