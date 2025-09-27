package main

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"time"

	gosteps "github.com/TanmoySG/go-steps"
	"github.com/rs/zerolog"
)

// Build represents a CI/CD build configuration
type Build struct {
	ID           string            `json:"id"`
	ProjectName  string            `json:"project_name"`
	Branch       string            `json:"branch"`
	CommitHash   string            `json:"commit_hash"`
	Environment  string            `json:"environment"`
	BuildType    string            `json:"build_type"` // feature, hotfix, release
	TestSuite    string            `json:"test_suite"` // unit, integration, e2e, all
	DeployTarget string            `json:"deploy_target"`
	Artifacts    []Artifact        `json:"artifacts"`
	Status       string            `json:"status"`
	StartTime    time.Time         `json:"start_time"`
	Metadata     map[string]string `json:"metadata"`
}

type Artifact struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // binary, docker, package
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Checksum string `json:"checksum"`
}

type TestResult struct {
	Suite    string  `json:"suite"`
	Passed   int     `json:"passed"`
	Failed   int     `json:"failed"`
	Skipped  int     `json:"skipped"`
	Duration string  `json:"duration"`
	Coverage float64 `json:"coverage"`
}

func main() {
	// Setup logging
	runLogFile, _ := os.OpenFile(
		"cicd-pipeline.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	output := zerolog.MultiLevelWriter(os.Stdout, runLogFile)
	logger := gosteps.NewGoStepsLogger(output, &gosteps.LoggerOpts{StepLoggingEnabled: true})

	// Initialize context with sample build
	ctx := gosteps.NewGoStepsContext()
	ctx.Use(logger)

	sampleBuild := Build{
		ID:           "BUILD-2025-001",
		ProjectName:  "microservice-api",
		Branch:       "feature/user-auth",
		CommitHash:   "a1b2c3d4e5f6",
		Environment:  "staging",
		BuildType:    "feature",
		TestSuite:    "all",
		DeployTarget: "k8s-staging",
		Status:       "pending",
		StartTime:    time.Now(),
		Artifacts:    []Artifact{},
		Metadata: map[string]string{
			"author":     "developer@company.com",
			"pr_number":  "123",
			"target_env": "staging",
		},
	}

	ctx.WithData(map[string]interface{}{
		"build":          sampleBuild,
		"testResults":    []TestResult{},
		"dockerImageTag": "",
		"deploymentID":   "",
		"notifications":  []string{},
	})

	// Define the CI/CD pipeline
	steps := gosteps.Steps{
		{
			Name:     "initializePipeline",
			Function: initializePipelineStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 2,
				RetrySleep:     1 * time.Second,
			},
		},
		{
			Name:     "checkoutCode",
			Function: checkoutCodeStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 3,
				ErrorPatternsToRetry: []regexp.Regexp{
					*regexp.MustCompile("git.*timeout"),
					*regexp.MustCompile("network.*error"),
				},
				RetrySleep: 2 * time.Second,
			},
			RollbackFunction: cleanupWorkspace,
		},
		{
			Name:     "buildApplication",
			Function: buildApplicationStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 2,
				RetrySleep:     1 * time.Second,
			},
			RollbackFunction: cleanupBuildArtifacts,
		},
		{
			Name:     "runTests",
			Function: runTestsStep,
			Branches: &gosteps.Branches{
				Resolver: testingResolver,
				Branches: []gosteps.Branch{
					{
						BranchName: "unitTestsOnly",
						Steps: gosteps.Steps{
							{
								Name:     "runUnitTests",
								Function: runUnitTestsStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 2,
									RetrySleep:     1 * time.Second,
								},
							},
						},
					},
					{
						BranchName: "integrationTests",
						Steps: gosteps.Steps{
							{
								Name:     "runUnitTests",
								Function: runUnitTestsStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 2,
								},
							},
							{
								Name:     "runIntegrationTests",
								Function: runIntegrationTestsStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 3,
									RetrySleep:     5 * time.Second,
								},
								RollbackFunction: cleanupTestEnvironment,
							},
						},
					},
					{
						BranchName: "fullTestSuite",
						Steps: gosteps.Steps{
							{
								Name:     "runUnitTests",
								Function: runUnitTestsStep,
							},
							{
								Name:     "runIntegrationTests",
								Function: runIntegrationTestsStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 2,
									RetrySleep:     3 * time.Second,
								},
								RollbackFunction: cleanupTestEnvironment,
							},
							{
								Name:     "runE2ETests",
								Function: runE2ETestsStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 2,
									RetrySleep:     10 * time.Second,
								},
								RollbackFunction: cleanupTestEnvironment,
							},
						},
					},
				},
			},
		},
		{
			Name:     "buildDockerImage",
			Function: buildDockerImageStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 3,
				RetrySleep:     2 * time.Second,
			},
			RollbackFunction: cleanupDockerImages,
		},
		{
			Name:     "securityScanning",
			Function: securityScanningStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 2,
				ErrorPatternsToRetry: []regexp.Regexp{
					*regexp.MustCompile("scanner.*unavailable"),
					*regexp.MustCompile("timeout.*scanning"),
				},
				RetrySleep: 5 * time.Second,
			},
		},
		{
			Name:     "deployApplication",
			Function: deployApplicationStep,
			Branches: &gosteps.Branches{
				Resolver: deploymentResolver,
				Branches: []gosteps.Branch{
					{
						BranchName: "stagingDeploy",
						Steps: gosteps.Steps{
							{
								Name:     "deployToStaging",
								Function: deployToStagingStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 3,
									RetrySleep:     5 * time.Second,
								},
								RollbackFunction: rollbackDeployment,
							},
							{
								Name:     "verifyStagingDeploy",
								Function: verifyStagingDeployStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 5,
									RetrySleep:     3 * time.Second,
								},
							},
						},
					},
					{
						BranchName: "productionDeploy",
						Steps: gosteps.Steps{
							{
								Name:     "deployToProduction",
								Function: deployToProductionStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 2,
									RetrySleep:     10 * time.Second,
								},
								RollbackFunction: rollbackDeployment,
							},
							{
								Name:     "verifyProductionDeploy",
								Function: verifyProductionDeployStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 3,
									RetrySleep:     5 * time.Second,
								},
							},
							{
								Name:     "runSmokeTests",
								Function: runSmokeTestsStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 2,
									RetrySleep:     5 * time.Second,
								},
							},
						},
					},
					{
						BranchName: "canaryDeploy",
						Steps: gosteps.Steps{
							{
								Name:             "deployCanary",
								Function:         deployCanaryStep,
								RollbackFunction: rollbackDeployment,
							},
							{
								Name:     "monitorCanary",
								Function: monitorCanaryStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 10,
									RetrySleep:     30 * time.Second,
								},
							},
							{
								Name:             "promoteCanary",
								Function:         promoteCanaryStep,
								RollbackFunction: rollbackDeployment,
							},
						},
					},
				},
			},
		},
		{
			Name:     "notifyTeam",
			Function: notifyTeamStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 3,
				RetrySleep:     2 * time.Second,
			},
		},
	}

	// Execute the CI/CD pipeline
	fmt.Println("🚀 Starting CI/CD Pipeline...")
	stepsProcessor := gosteps.NewStepsProcessor(steps)
	stepsProcessor.Execute(ctx)

	// Print final build status
	finalBuild := ctx.GetData("build").(Build)
	fmt.Printf("\n🎯 CI/CD Pipeline Complete!\n")
	fmt.Printf("Build ID: %s\n", finalBuild.ID)
	fmt.Printf("Project: %s\n", finalBuild.ProjectName)
	fmt.Printf("Branch: %s\n", finalBuild.Branch)
	fmt.Printf("Status: %s\n", finalBuild.Status)
	fmt.Printf("Duration: %v\n", time.Since(finalBuild.StartTime).Round(time.Second))
}

// Step Functions

func initializePipelineStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Initializing CI/CD pipeline...")

	build := ctx.GetData("build").(Build)
	build.Status = "initializing"
	build.StartTime = time.Now()

	ctx.SetData("build", build)
	ctx.SetData("pipelineStartTime", time.Now())

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Pipeline initialized for %s@%s", build.ProjectName, build.Branch))
}

func checkoutCodeStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Checking out source code...")

	build := ctx.GetData("build").(Build)

	// Simulate git checkout with potential network issues
	rand.Seed(time.Now().UnixNano())
	if rand.Float32() < 0.1 {
		return gosteps.MarkStateError().WithError(fmt.Errorf("git timeout while checking out %s", build.Branch))
	}

	build.Status = "code_checked_out"
	ctx.SetData("build", build)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Code checked out from %s@%s", build.Branch, build.CommitHash))
}

func buildApplicationStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Building application artifacts...")

	build := ctx.GetData("build").(Build)

	// Simulate build process
	time.Sleep(100 * time.Millisecond) // Simulate build time

	// Create build artifacts
	artifacts := []Artifact{
		{
			Name:     fmt.Sprintf("%s-binary", build.ProjectName),
			Type:     "binary",
			Path:     fmt.Sprintf("/build/%s", build.ProjectName),
			Size:     1024000, // 1MB
			Checksum: "sha256:abc123def456",
		},
	}

	build.Artifacts = artifacts
	build.Status = "built"
	ctx.SetData("build", build)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Application built successfully, %d artifacts generated", len(artifacts)))
}

func runTestsStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Determining test strategy...")

	build := ctx.GetData("build").(Build)
	build.Status = "testing"
	ctx.SetData("build", build)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Test strategy determined for %s build", build.BuildType))
}

func testingResolver(ctx gosteps.GoStepsCtx) gosteps.BranchName {
	build := ctx.GetData("build").(Build)

	switch build.TestSuite {
	case "unit":
		ctx.Log("Running unit tests only")
		return gosteps.BranchName("unitTestsOnly")
	case "integration":
		ctx.Log("Running unit and integration tests")
		return gosteps.BranchName("integrationTests")
	case "all", "e2e":
		ctx.Log("Running full test suite including E2E tests")
		return gosteps.BranchName("fullTestSuite")
	default:
		ctx.Log("Defaulting to unit tests")
		return gosteps.BranchName("unitTestsOnly")
	}
}

func runUnitTestsStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Running unit tests...")

	// Simulate unit test execution
	time.Sleep(50 * time.Millisecond)

	testResult := TestResult{
		Suite:    "unit",
		Passed:   45,
		Failed:   2,
		Skipped:  1,
		Duration: "2.5s",
		Coverage: 85.5,
	}

	// Add to test results
	testResults := ctx.GetData("testResults").([]TestResult)
	testResults = append(testResults, testResult)
	ctx.SetData("testResults", testResults)

	if testResult.Failed > 5 {
		return gosteps.MarkStateFailed().WithMessage("Too many unit test failures")
	}

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Unit tests: %d passed, %d failed, %.1f%% coverage",
			testResult.Passed, testResult.Failed, testResult.Coverage))
}

func runIntegrationTestsStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Running integration tests...")

	// Simulate integration test execution with potential flakiness
	time.Sleep(200 * time.Millisecond)

	if rand.Float32() < 0.15 {
		return gosteps.MarkStatePending().WithMessage("Integration test environment temporarily unavailable")
	}

	testResult := TestResult{
		Suite:    "integration",
		Passed:   28,
		Failed:   1,
		Skipped:  0,
		Duration: "45s",
		Coverage: 72.0,
	}

	// Add to test results
	testResults := ctx.GetData("testResults").([]TestResult)
	testResults = append(testResults, testResult)
	ctx.SetData("testResults", testResults)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Integration tests: %d passed, %d failed",
			testResult.Passed, testResult.Failed))
}

func runE2ETestsStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Running end-to-end tests...")

	// E2E tests take longer and are more prone to failures
	time.Sleep(500 * time.Millisecond)

	if rand.Float32() < 0.2 {
		return gosteps.MarkStatePending().WithMessage("E2E test environment setup in progress")
	}

	testResult := TestResult{
		Suite:    "e2e",
		Passed:   15,
		Failed:   0,
		Skipped:  2,
		Duration: "3m45s",
		Coverage: 95.0,
	}

	// Add to test results
	testResults := ctx.GetData("testResults").([]TestResult)
	testResults = append(testResults, testResult)
	ctx.SetData("testResults", testResults)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("E2E tests: %d passed, %d failed",
			testResult.Passed, testResult.Failed))
}

func buildDockerImageStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Building Docker image...")

	build := ctx.GetData("build").(Build)

	// Generate docker image tag
	imageTag := fmt.Sprintf("%s:%s-%s", build.ProjectName, build.Branch, build.CommitHash[:8])
	ctx.SetData("dockerImageTag", imageTag)

	// Add docker artifact
	dockerArtifact := Artifact{
		Name:     fmt.Sprintf("%s-docker", build.ProjectName),
		Type:     "docker",
		Path:     imageTag,
		Size:     150000000, // 150MB
		Checksum: "sha256:docker123abc456",
	}

	build.Artifacts = append(build.Artifacts, dockerArtifact)
	build.Status = "docker_built"
	ctx.SetData("build", build)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Docker image built: %s", imageTag))
}

func securityScanningStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Running security scans...")

	// Simulate security scanning
	time.Sleep(300 * time.Millisecond)

	if rand.Float32() < 0.1 {
		return gosteps.MarkStateError().WithError(fmt.Errorf("scanner unavailable - security service timeout"))
	}

	// Simulate security scan results
	vulnCount := rand.Intn(3) // 0-2 vulnerabilities

	if vulnCount > 1 {
		return gosteps.MarkStateFailed().
			WithMessage(fmt.Sprintf("Security scan failed: %d high-severity vulnerabilities found", vulnCount))
	}

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Security scan passed: %d vulnerabilities found", vulnCount))
}

func deployApplicationStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Determining deployment strategy...")

	build := ctx.GetData("build").(Build)
	build.Status = "deploying"
	ctx.SetData("build", build)

	return gosteps.MarkStateComplete().WithMessage("Deployment strategy determined")
}

func deploymentResolver(ctx gosteps.GoStepsCtx) gosteps.BranchName {
	build := ctx.GetData("build").(Build)

	switch {
	case build.Environment == "production" && build.BuildType == "release":
		ctx.Log("Production release - using canary deployment")
		return gosteps.BranchName("canaryDeploy")
	case build.Environment == "production":
		ctx.Log("Production deployment - using standard strategy")
		return gosteps.BranchName("productionDeploy")
	default:
		ctx.Log("Non-production deployment - using staging strategy")
		return gosteps.BranchName("stagingDeploy")
	}
}

func deployToStagingStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Deploying to staging environment...")

	build := ctx.GetData("build").(Build)
	imageTag := ctx.GetData("dockerImageTag").(string)

	deploymentID := fmt.Sprintf("staging-deploy-%d", time.Now().UnixNano())
	ctx.SetData("deploymentID", deploymentID)

	build.Status = "deployed_staging"
	ctx.SetData("build", build)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Deployed %s to staging. Deployment ID: %s", imageTag, deploymentID))
}

func verifyStagingDeployStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Verifying staging deployment...")

	deploymentID := ctx.GetData("deploymentID").(string)

	// Simulate deployment verification with potential delays
	if rand.Float32() < 0.2 {
		return gosteps.MarkStatePending().WithMessage("Staging environment starting up...")
	}

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Staging deployment %s verified successfully", deploymentID))
}

func deployToProductionStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Deploying to production environment...")

	build := ctx.GetData("build").(Build)
	imageTag := ctx.GetData("dockerImageTag").(string)

	deploymentID := fmt.Sprintf("prod-deploy-%d", time.Now().UnixNano())
	ctx.SetData("deploymentID", deploymentID)

	build.Status = "deployed_production"
	ctx.SetData("build", build)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Deployed %s to production. Deployment ID: %s", imageTag, deploymentID))
}

func verifyProductionDeployStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Verifying production deployment...")

	deploymentID := ctx.GetData("deploymentID").(string)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Production deployment %s verified successfully", deploymentID))
}

func runSmokeTestsStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Running production smoke tests...")

	// Simulate smoke tests
	time.Sleep(100 * time.Millisecond)

	smokeTestResult := TestResult{
		Suite:    "smoke",
		Passed:   8,
		Failed:   0,
		Skipped:  0,
		Duration: "30s",
		Coverage: 100.0,
	}

	testResults := ctx.GetData("testResults").([]TestResult)
	testResults = append(testResults, smokeTestResult)
	ctx.SetData("testResults", testResults)

	return gosteps.MarkStateComplete().
		WithMessage("Production smoke tests passed")
}

func deployCanaryStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Deploying canary release...")

	build := ctx.GetData("build").(Build)
	imageTag := ctx.GetData("dockerImageTag").(string)

	deploymentID := fmt.Sprintf("canary-deploy-%d", time.Now().UnixNano())
	ctx.SetData("deploymentID", deploymentID)

	build.Status = "canary_deployed"
	ctx.SetData("build", build)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Canary deployment %s started with %s", deploymentID, imageTag))
}

func monitorCanaryStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Monitoring canary deployment metrics...")

	// Simulate canary monitoring
	time.Sleep(100 * time.Millisecond)

	// Simulate metrics - error rate, latency, etc.
	errorRate := rand.Float64() * 0.05 // 0-5% error rate
	if errorRate > 0.02 {              // 2% threshold
		return gosteps.MarkStateFailed().
			WithMessage(fmt.Sprintf("Canary metrics indicate issues: %.2f%% error rate", errorRate*100))
	}

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Canary metrics healthy: %.2f%% error rate", errorRate*100))
}

func promoteCanaryStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Promoting canary to full production...")

	build := ctx.GetData("build").(Build)
	build.Status = "deployed_production"
	ctx.SetData("build", build)

	return gosteps.MarkStateComplete().WithMessage("Canary promoted to full production")
}

func notifyTeamStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Sending pipeline completion notifications...")

	build := ctx.GetData("build").(Build)
	testResults := ctx.GetData("testResults").([]TestResult)

	// Generate notification summary
	totalTests := 0
	totalPassed := 0
	totalFailed := 0

	for _, result := range testResults {
		totalTests += result.Passed + result.Failed + result.Skipped
		totalPassed += result.Passed
		totalFailed += result.Failed
	}

	notifications := []string{
		fmt.Sprintf("Build %s completed with status: %s", build.ID, build.Status),
		fmt.Sprintf("Tests: %d passed, %d failed out of %d total", totalPassed, totalFailed, totalTests),
		fmt.Sprintf("Artifacts: %d generated", len(build.Artifacts)),
	}

	ctx.SetData("notifications", notifications)

	if build.Status == "deployed_production" {
		build.Status = "completed"
	} else {
		build.Status = "completed_with_warnings"
	}
	ctx.SetData("build", build)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Team notified of build completion: %s", build.Status))
}

// Rollback Functions

func cleanupWorkspace(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Cleaning up workspace...")

	// Simulate workspace cleanup
	build := ctx.GetData("build").(Build)
	build.Status = "workspace_cleaned"
	ctx.SetData("build", build)

	message := "Workspace cleaned up successfully"
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}

func cleanupBuildArtifacts(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Cleaning up build artifacts...")

	build := ctx.GetData("build").(Build)
	build.Artifacts = []Artifact{}
	ctx.SetData("build", build)

	message := "Build artifacts cleaned up"
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}

func cleanupTestEnvironment(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Cleaning up test environment...")

	// Reset test results
	ctx.SetData("testResults", []TestResult{})

	message := "Test environment cleaned up"
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}

func cleanupDockerImages(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Cleaning up Docker images...")

	imageTag := ctx.GetData("dockerImageTag").(string)
	if imageTag != "" {
		ctx.Log(fmt.Sprintf("Removing Docker image: %s", imageTag))
		ctx.SetData("dockerImageTag", "")
	}

	message := "Docker images cleaned up"
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}

func rollbackDeployment(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Rolling back deployment...")

	build := ctx.GetData("build").(Build)
	deploymentID := ctx.GetData("deploymentID").(string)

	if deploymentID != "" {
		ctx.Log(fmt.Sprintf("Rolling back deployment: %s", deploymentID))

		// Simulate rollback process
		build.Status = "deployment_rolled_back"
		ctx.SetData("build", build)
		ctx.SetData("deploymentID", "")

		message := fmt.Sprintf("Deployment %s rolled back successfully", deploymentID)
		return gosteps.RollbackResult{
			RollbackState:   gosteps.RollbackStateSuccess,
			RollbackMessage: &message,
		}
	}

	message := "No deployment to rollback"
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}
