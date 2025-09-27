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

// ProvisioningJob represents a cloud infrastructure provisioning job
type ProvisioningJob struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	CloudProvider string                 `json:"cloud_provider"`
	Region        string                 `json:"region"`
	Environment   string                 `json:"environment"`
	Resources     []CloudResource        `json:"resources"`
	Dependencies  []Dependency           `json:"dependencies"`
	Status        string                 `json:"status"`
	StartTime     time.Time              `json:"start_time"`
	Progress      ProvisioningProgress   `json:"progress"`
	Configuration map[string]interface{} `json:"configuration"`
}

type CloudResource struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Type           string                 `json:"type"` // vpc, subnet, instance, database, loadbalancer
	Specifications map[string]interface{} `json:"specifications"`
	Status         string                 `json:"status"`
	Dependencies   []string               `json:"dependencies"`
	CreatedAt      *time.Time             `json:"created_at,omitempty"`
	ResourceARN    string                 `json:"resource_arn,omitempty"`
}

type Dependency struct {
	ResourceID     string   `json:"resource_id"`
	DependsOn      []string `json:"depends_on"`
	DependencyType string   `json:"dependency_type"` // hard, soft
}

type ProvisioningProgress struct {
	ResourcesTotal      int               `json:"resources_total"`
	ResourcesCompleted  int               `json:"resources_completed"`
	ResourcesFailed     int               `json:"resources_failed"`
	CurrentPhase        string            `json:"current_phase"`
	EstimatedCompletion time.Time         `json:"estimated_completion"`
	ResourceStatus      map[string]string `json:"resource_status"`
}

func main() {
	// Setup logging
	runLogFile, _ := os.OpenFile(
		"cloud-provisioning.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	output := zerolog.MultiLevelWriter(os.Stdout, runLogFile)
	logger := gosteps.NewGoStepsLogger(output, &gosteps.LoggerOpts{StepLoggingEnabled: true})

	// Initialize context with sample provisioning job
	ctx := gosteps.NewGoStepsContext()
	ctx.Use(logger)

	sampleJob := ProvisioningJob{
		ID:            "PROVISION-2025-001",
		Name:          "Production Web Application Infrastructure",
		CloudProvider: "aws",
		Region:        "us-west-2",
		Environment:   "production",
		Resources: []CloudResource{
			{
				ID:   "vpc-1",
				Name: "main-vpc",
				Type: "vpc",
				Specifications: map[string]interface{}{
					"cidr_block":           "10.0.0.0/16",
					"enable_dns_hostnames": true,
					"enable_dns_support":   true,
				},
				Status:       "pending",
				Dependencies: []string{},
			},
			{
				ID:   "subnet-1",
				Name: "public-subnet-1",
				Type: "subnet",
				Specifications: map[string]interface{}{
					"cidr_block":        "10.0.1.0/24",
					"availability_zone": "us-west-2a",
					"public":            true,
				},
				Status:       "pending",
				Dependencies: []string{"vpc-1"},
			},
			{
				ID:   "subnet-2",
				Name: "private-subnet-1",
				Type: "subnet",
				Specifications: map[string]interface{}{
					"cidr_block":        "10.0.2.0/24",
					"availability_zone": "us-west-2b",
					"public":            false,
				},
				Status:       "pending",
				Dependencies: []string{"vpc-1"},
			},
			{
				ID:   "db-1",
				Name: "main-database",
				Type: "database",
				Specifications: map[string]interface{}{
					"engine":            "postgresql",
					"instance_class":    "db.t3.medium",
					"allocated_storage": 100,
					"multi_az":          true,
				},
				Status:       "pending",
				Dependencies: []string{"subnet-2"},
			},
			{
				ID:   "app-1",
				Name: "web-application",
				Type: "instance",
				Specifications: map[string]interface{}{
					"instance_type": "t3.large",
					"ami_id":        "ami-0c55b159cbfafe1b0",
					"min_size":      2,
					"max_size":      10,
					"desired_size":  3,
				},
				Status:       "pending",
				Dependencies: []string{"subnet-1", "db-1"},
			},
			{
				ID:   "lb-1",
				Name: "application-loadbalancer",
				Type: "loadbalancer",
				Specifications: map[string]interface{}{
					"type":    "application",
					"scheme":  "internet-facing",
					"subnets": []string{"subnet-1"},
				},
				Status:       "pending",
				Dependencies: []string{"app-1"},
			},
		},
		Dependencies: []Dependency{
			{ResourceID: "subnet-1", DependsOn: []string{"vpc-1"}, DependencyType: "hard"},
			{ResourceID: "subnet-2", DependsOn: []string{"vpc-1"}, DependencyType: "hard"},
			{ResourceID: "db-1", DependsOn: []string{"subnet-2"}, DependencyType: "hard"},
			{ResourceID: "app-1", DependsOn: []string{"subnet-1", "db-1"}, DependencyType: "hard"},
			{ResourceID: "lb-1", DependsOn: []string{"app-1"}, DependencyType: "soft"},
		},
		Status:    "pending",
		StartTime: time.Now(),
		Progress: ProvisioningProgress{
			ResourcesTotal:     6,
			ResourcesCompleted: 0,
			ResourcesFailed:    0,
			CurrentPhase:       "initialization",
			ResourceStatus:     make(map[string]string),
		},
		Configuration: map[string]interface{}{
			"timeout":         30 * time.Minute,
			"retry_attempts":  3,
			"parallel_limit":  3,
			"destroy_on_fail": true,
		},
	}

	ctx.WithData(map[string]interface{}{
		"job":                  sampleJob,
		"provisionedResources": make(map[string]CloudResource),
		"failedResources":      []string{},
		"rollbackRequired":     false,
	})

	// Define the cloud provisioning workflow
	steps := gosteps.Steps{
		{
			Name:     "validateConfiguration",
			Function: validateConfigurationStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 2,
				RetrySleep:     1 * time.Second,
			},
		},
		{
			Name:     "authenticateProvider",
			Function: authenticateProviderStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 3,
				ErrorPatternsToRetry: []regexp.Regexp{
					*regexp.MustCompile("authentication.*timeout"),
					*regexp.MustCompile("credential.*expired"),
				},
				RetrySleep: 2 * time.Second,
			},
		},
		{
			Name:     "analyzeDependencies",
			Function: analyzeDependenciesStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 2,
			},
		},
		{
			Name:     "provisionResources",
			Function: provisionResourcesStep,
			Branches: &gosteps.Branches{
				Resolver: provisioningStrategyResolver,
				Branches: []gosteps.Branch{
					{
						BranchName: "networkFirstStrategy",
						Steps: gosteps.Steps{
							{
								Name:     "provisionNetworking",
								Function: provisionNetworkingStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 3,
									RetrySleep:     5 * time.Second,
								},
								RollbackFunction: destroyNetworkingResources,
							},
							{
								Name:     "provisionCompute",
								Function: provisionComputeStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 3,
									RetrySleep:     10 * time.Second,
								},
								RollbackFunction: destroyComputeResources,
							},
							{
								Name:     "provisionDatabase",
								Function: provisionDatabaseStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 2,
									RetrySleep:     15 * time.Second,
								},
								RollbackFunction: destroyDatabaseResources,
							},
							{
								Name:     "provisionLoadBalancer",
								Function: provisionLoadBalancerStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 2,
									RetrySleep:     5 * time.Second,
								},
								RollbackFunction: destroyLoadBalancerResources,
							},
						},
					},
					{
						BranchName: "parallelStrategy",
						Steps: gosteps.Steps{
							{
								Name:     "provisionIndependentResources",
								Function: provisionIndependentResourcesStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 3,
									RetrySleep:     5 * time.Second,
								},
								RollbackFunction: destroyProvisionedResources,
							},
							{
								Name:     "provisionDependentResources",
								Function: provisionDependentResourcesStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 3,
									RetrySleep:     10 * time.Second,
								},
								RollbackFunction: destroyProvisionedResources,
							},
						},
					},
				},
			},
		},
		{
			Name:     "configureResources",
			Function: configureResourcesStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 3,
				RetrySleep:     5 * time.Second,
			},
			RollbackFunction: resetResourceConfiguration,
		},
		{
			Name:     "validateDeployment",
			Function: validateDeploymentStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 5,
				RetrySleep:     10 * time.Second,
			},
		},
		{
			Name:     "generateOutputs",
			Function: generateOutputsStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 2,
			},
		},
	}

	// Execute the cloud provisioning workflow
	fmt.Println("☁️ Starting Cloud Resource Provisioning...")
	stepsProcessor := gosteps.NewStepsProcessor(steps)
	stepsProcessor.Execute(ctx)

	// Print final provisioning status
	finalJob := ctx.GetData("job").(ProvisioningJob)
	fmt.Printf("\n🎯 Cloud Provisioning Complete!\n")
	fmt.Printf("Job ID: %s\n", finalJob.ID)
	fmt.Printf("Status: %s\n", finalJob.Status)
	fmt.Printf("Resources: %d total, %d completed, %d failed\n",
		finalJob.Progress.ResourcesTotal,
		finalJob.Progress.ResourcesCompleted,
		finalJob.Progress.ResourcesFailed)
	fmt.Printf("Duration: %v\n", time.Since(finalJob.StartTime).Round(time.Second))
}

// Step Functions

func validateConfigurationStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Validating provisioning configuration...")

	job := ctx.GetData("job").(ProvisioningJob)

	// Validate basic configuration
	if job.CloudProvider == "" {
		return gosteps.MarkStateFailed().WithMessage("Cloud provider must be specified")
	}

	if job.Region == "" {
		return gosteps.MarkStateFailed().WithMessage("Region must be specified")
	}

	if len(job.Resources) == 0 {
		return gosteps.MarkStateFailed().WithMessage("At least one resource must be specified")
	}

	// Validate resource specifications
	for _, resource := range job.Resources {
		if resource.Name == "" {
			return gosteps.MarkStateFailed().
				WithMessage(fmt.Sprintf("Resource %s must have a name", resource.ID))
		}

		if resource.Type == "" {
			return gosteps.MarkStateFailed().
				WithMessage(fmt.Sprintf("Resource %s must have a type", resource.ID))
		}
	}

	job.Status = "configuration_validated"
	job.Progress.CurrentPhase = "configuration_validation_complete"
	ctx.SetData("job", job)

	return gosteps.MarkStateComplete().WithMessage("Configuration validation successful")
}

func authenticateProviderStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Authenticating with cloud provider...")

	job := ctx.GetData("job").(ProvisioningJob)

	// Simulate authentication with potential failures
	if rand.Float32() < 0.1 {
		return gosteps.MarkStateError().WithError(fmt.Errorf("authentication timeout with %s", job.CloudProvider))
	}

	// Simulate credential validation
	time.Sleep(100 * time.Millisecond)

	authToken := fmt.Sprintf("auth-token-%d", time.Now().UnixNano())

	job.Status = "authenticated"
	job.Progress.CurrentPhase = "authentication_complete"
	ctx.SetData("job", job)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Successfully authenticated with %s", job.CloudProvider)).
		WithData(map[string]interface{}{
			"authToken": authToken,
		})
}

func analyzeDependenciesStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Analyzing resource dependencies...")

	job := ctx.GetData("job").(ProvisioningJob)

	// Build dependency graph
	dependencyMap := make(map[string][]string)

	for _, dep := range job.Dependencies {
		dependencyMap[dep.ResourceID] = dep.DependsOn
	}

	// Validate dependency graph for cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for _, resource := range job.Resources {
		if hasCycle(resource.ID, dependencyMap, visited, recStack) {
			return gosteps.MarkStateFailed().
				WithMessage(fmt.Sprintf("Circular dependency detected involving resource %s", resource.ID))
		}
	}

	// Calculate provisioning order
	provisioningOrder := topologicalSort(job.Resources, dependencyMap)

	job.Status = "dependencies_analyzed"
	job.Progress.CurrentPhase = "dependency_analysis_complete"
	ctx.SetData("job", job)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Dependency analysis complete. Provisioning order: %v", provisioningOrder)).
		WithData(map[string]interface{}{
			"provisioningOrder": provisioningOrder,
			"dependencyMap":     dependencyMap,
		})
}

func provisionResourcesStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Determining provisioning strategy...")

	job := ctx.GetData("job").(ProvisioningJob)
	job.Status = "provisioning"
	job.Progress.CurrentPhase = "resource_provisioning"
	ctx.SetData("job", job)

	return gosteps.MarkStateComplete().WithMessage("Provisioning strategy determined")
}

func provisioningStrategyResolver(ctx gosteps.GoStepsCtx) gosteps.BranchName {
	job := ctx.GetData("job").(ProvisioningJob)

	// Determine strategy based on resource count and complexity
	if len(job.Resources) > 5 && job.Environment == "production" {
		ctx.Log("Large deployment - using network-first strategy for stability")
		return gosteps.BranchName("networkFirstStrategy")
	} else {
		ctx.Log("Standard deployment - using parallel strategy for speed")
		return gosteps.BranchName("parallelStrategy")
	}
}

func provisionNetworkingStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Provisioning networking resources...")

	job := ctx.GetData("job").(ProvisioningJob)
	provisionedResources := ctx.GetData("provisionedResources").(map[string]CloudResource)

	// Provision VPC and subnets first
	networkResources := []string{"vpc", "subnet"}

	for _, resource := range job.Resources {
		for _, networkType := range networkResources {
			if resource.Type == networkType {
				if err := provisionResource(ctx, resource); err != nil {
					return gosteps.MarkStateError().WithError(err)
				}

				resource.Status = "provisioned"
				resource.CreatedAt = &[]time.Time{time.Now()}[0]
				resource.ResourceARN = fmt.Sprintf("arn:aws:%s:%s:account:%s",
					resource.Type, job.Region, resource.ID)

				provisionedResources[resource.ID] = resource
				job.Progress.ResourcesCompleted++
				job.Progress.ResourceStatus[resource.ID] = "provisioned"
			}
		}
	}

	ctx.SetData("job", job)
	ctx.SetData("provisionedResources", provisionedResources)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Networking resources provisioned: %d completed", job.Progress.ResourcesCompleted))
}

func provisionComputeStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Provisioning compute resources...")

	job := ctx.GetData("job").(ProvisioningJob)
	provisionedResources := ctx.GetData("provisionedResources").(map[string]CloudResource)

	// Provision instances and auto-scaling groups
	for _, resource := range job.Resources {
		if resource.Type == "instance" {
			if err := provisionResource(ctx, resource); err != nil {
				return gosteps.MarkStateError().WithError(err)
			}

			resource.Status = "provisioned"
			resource.CreatedAt = &[]time.Time{time.Now()}[0]
			resource.ResourceARN = fmt.Sprintf("arn:aws:ec2:%s:account:instance/%s",
				job.Region, resource.ID)

			provisionedResources[resource.ID] = resource
			job.Progress.ResourcesCompleted++
			job.Progress.ResourceStatus[resource.ID] = "provisioned"
		}
	}

	ctx.SetData("job", job)
	ctx.SetData("provisionedResources", provisionedResources)

	return gosteps.MarkStateComplete().WithMessage("Compute resources provisioned successfully")
}

func provisionDatabaseStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Provisioning database resources...")

	job := ctx.GetData("job").(ProvisioningJob)
	provisionedResources := ctx.GetData("provisionedResources").(map[string]CloudResource)

	// Provision databases
	for _, resource := range job.Resources {
		if resource.Type == "database" {
			// Database provisioning takes longer
			if err := provisionResource(ctx, resource); err != nil {
				return gosteps.MarkStateError().WithError(err)
			}

			// Simulate longer database setup time
			time.Sleep(200 * time.Millisecond)

			resource.Status = "provisioned"
			resource.CreatedAt = &[]time.Time{time.Now()}[0]
			resource.ResourceARN = fmt.Sprintf("arn:aws:rds:%s:account:db:%s",
				job.Region, resource.ID)

			provisionedResources[resource.ID] = resource
			job.Progress.ResourcesCompleted++
			job.Progress.ResourceStatus[resource.ID] = "provisioned"
		}
	}

	ctx.SetData("job", job)
	ctx.SetData("provisionedResources", provisionedResources)

	return gosteps.MarkStateComplete().WithMessage("Database resources provisioned successfully")
}

func provisionLoadBalancerStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Provisioning load balancer resources...")

	job := ctx.GetData("job").(ProvisioningJob)
	provisionedResources := ctx.GetData("provisionedResources").(map[string]CloudResource)

	// Provision load balancers
	for _, resource := range job.Resources {
		if resource.Type == "loadbalancer" {
			if err := provisionResource(ctx, resource); err != nil {
				return gosteps.MarkStateError().WithError(err)
			}

			resource.Status = "provisioned"
			resource.CreatedAt = &[]time.Time{time.Now()}[0]
			resource.ResourceARN = fmt.Sprintf("arn:aws:elasticloadbalancing:%s:account:loadbalancer/%s",
				job.Region, resource.ID)

			provisionedResources[resource.ID] = resource
			job.Progress.ResourcesCompleted++
			job.Progress.ResourceStatus[resource.ID] = "provisioned"
		}
	}

	ctx.SetData("job", job)
	ctx.SetData("provisionedResources", provisionedResources)

	return gosteps.MarkStateComplete().WithMessage("Load balancer resources provisioned successfully")
}

func provisionIndependentResourcesStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Provisioning independent resources in parallel...")

	job := ctx.GetData("job").(ProvisioningJob)
	provisionedResources := ctx.GetData("provisionedResources").(map[string]CloudResource)

	// Find resources with no dependencies
	independentResources := []CloudResource{}
	for _, resource := range job.Resources {
		if len(resource.Dependencies) == 0 {
			independentResources = append(independentResources, resource)
		}
	}

	// Provision independent resources
	for _, resource := range independentResources {
		if err := provisionResource(ctx, resource); err != nil {
			return gosteps.MarkStateError().WithError(err)
		}

		resource.Status = "provisioned"
		resource.CreatedAt = &[]time.Time{time.Now()}[0]
		resource.ResourceARN = generateResourceARN(job.Region, resource)

		provisionedResources[resource.ID] = resource
		job.Progress.ResourcesCompleted++
		job.Progress.ResourceStatus[resource.ID] = "provisioned"
	}

	ctx.SetData("job", job)
	ctx.SetData("provisionedResources", provisionedResources)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Independent resources provisioned: %d completed", len(independentResources)))
}

func provisionDependentResourcesStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Provisioning dependent resources...")

	job := ctx.GetData("job").(ProvisioningJob)
	provisionedResources := ctx.GetData("provisionedResources").(map[string]CloudResource)

	// Find resources with dependencies that are now satisfied
	for _, resource := range job.Resources {
		if len(resource.Dependencies) > 0 && resource.Status != "provisioned" {
			// Check if all dependencies are satisfied
			allDependenciesSatisfied := true
			for _, depID := range resource.Dependencies {
				if _, exists := provisionedResources[depID]; !exists {
					allDependenciesSatisfied = false
					break
				}
			}

			if allDependenciesSatisfied {
				if err := provisionResource(ctx, resource); err != nil {
					return gosteps.MarkStateError().WithError(err)
				}

				resource.Status = "provisioned"
				resource.CreatedAt = &[]time.Time{time.Now()}[0]
				resource.ResourceARN = generateResourceARN(job.Region, resource)

				provisionedResources[resource.ID] = resource
				job.Progress.ResourcesCompleted++
				job.Progress.ResourceStatus[resource.ID] = "provisioned"
			}
		}
	}

	ctx.SetData("job", job)
	ctx.SetData("provisionedResources", provisionedResources)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Dependent resources provisioned: %d total completed", job.Progress.ResourcesCompleted))
}

func configureResourcesStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Configuring provisioned resources...")

	job := ctx.GetData("job").(ProvisioningJob)
	provisionedResources := ctx.GetData("provisionedResources").(map[string]CloudResource)

	// Configure each provisioned resource
	configurationCount := 0
	for resourceID, resource := range provisionedResources {
		ctx.Log(fmt.Sprintf("Configuring resource: %s (%s)", resource.Name, resource.Type))

		// Simulate configuration with potential failures
		if rand.Float32() < 0.05 {
			return gosteps.MarkStateError().WithError(fmt.Errorf("configuration failed for resource %s", resourceID))
		}

		time.Sleep(30 * time.Millisecond)
		configurationCount++
	}

	job.Status = "configured"
	job.Progress.CurrentPhase = "resource_configuration_complete"
	ctx.SetData("job", job)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Resource configuration completed: %d resources configured", configurationCount))
}

func validateDeploymentStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Validating deployment...")

	job := ctx.GetData("job").(ProvisioningJob)
	provisionedResources := ctx.GetData("provisionedResources").(map[string]CloudResource)

	// Validate that all resources are healthy and accessible
	validationChecks := []string{
		"connectivity_check",
		"health_check",
		"security_group_validation",
		"dns_resolution_check",
		"load_balancer_targets",
	}

	for _, check := range validationChecks {
		ctx.Log(fmt.Sprintf("Running validation: %s", check))
		time.Sleep(20 * time.Millisecond)

		// Simulate potential validation failures
		if rand.Float32() < 0.1 {
			return gosteps.MarkStatePending().WithMessage(fmt.Sprintf("Validation %s pending - resources still initializing", check))
		}
	}

	// Verify all expected resources are provisioned
	if len(provisionedResources) != len(job.Resources) {
		return gosteps.MarkStateFailed().
			WithMessage(fmt.Sprintf("Resource count mismatch: expected %d, got %d",
				len(job.Resources), len(provisionedResources)))
	}

	job.Status = "validated"
	job.Progress.CurrentPhase = "deployment_validation_complete"
	ctx.SetData("job", job)

	return gosteps.MarkStateComplete().WithMessage("Deployment validation successful")
}

func generateOutputsStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Generating deployment outputs...")

	job := ctx.GetData("job").(ProvisioningJob)
	provisionedResources := ctx.GetData("provisionedResources").(map[string]CloudResource)

	// Generate outputs for external consumption
	outputs := map[string]interface{}{
		"deployment_id":     job.ID,
		"region":            job.Region,
		"environment":       job.Environment,
		"total_resources":   len(provisionedResources),
		"provisioning_time": time.Since(job.StartTime).String(),
		"resource_arns":     make(map[string]string),
		"endpoints":         make(map[string]string),
	}

	// Extract resource ARNs and endpoints
	resourceARNs := outputs["resource_arns"].(map[string]string)
	endpoints := outputs["endpoints"].(map[string]string)

	for resourceID, resource := range provisionedResources {
		resourceARNs[resourceID] = resource.ResourceARN

		// Generate endpoints for applicable resources
		switch resource.Type {
		case "database":
			endpoints[resourceID] = fmt.Sprintf("%s.%s.rds.amazonaws.com:5432",
				resource.ID, job.Region)
		case "loadbalancer":
			endpoints[resourceID] = fmt.Sprintf("%s-lb-%s.elb.amazonaws.com",
				resource.Name, job.Region)
		}
	}

	job.Status = "completed"
	job.Progress.CurrentPhase = "deployment_complete"
	job.Progress.EstimatedCompletion = time.Now()
	ctx.SetData("job", job)

	return gosteps.MarkStateComplete().
		WithMessage("Deployment outputs generated successfully").
		WithData(map[string]interface{}{
			"outputs": outputs,
		})
}

// Helper Functions

func provisionResource(ctx gosteps.GoStepsCtx, resource CloudResource) error {
	ctx.Log(fmt.Sprintf("Provisioning %s: %s", resource.Type, resource.Name))

	// Simulate provisioning time based on resource type
	switch resource.Type {
	case "vpc":
		time.Sleep(50 * time.Millisecond)
	case "subnet":
		time.Sleep(30 * time.Millisecond)
	case "instance":
		time.Sleep(100 * time.Millisecond)
	case "database":
		time.Sleep(200 * time.Millisecond)
	case "loadbalancer":
		time.Sleep(80 * time.Millisecond)
	default:
		time.Sleep(40 * time.Millisecond)
	}

	// Simulate random provisioning failures
	if rand.Float32() < 0.05 {
		return fmt.Errorf("provisioning failed for %s: resource limit exceeded", resource.Name)
	}

	return nil
}

func generateResourceARN(region string, resource CloudResource) string {
	switch resource.Type {
	case "vpc":
		return fmt.Sprintf("arn:aws:ec2:%s:account:vpc/%s", region, resource.ID)
	case "subnet":
		return fmt.Sprintf("arn:aws:ec2:%s:account:subnet/%s", region, resource.ID)
	case "instance":
		return fmt.Sprintf("arn:aws:ec2:%s:account:instance/%s", region, resource.ID)
	case "database":
		return fmt.Sprintf("arn:aws:rds:%s:account:db:%s", region, resource.ID)
	case "loadbalancer":
		return fmt.Sprintf("arn:aws:elasticloadbalancing:%s:account:loadbalancer/%s", region, resource.ID)
	default:
		return fmt.Sprintf("arn:aws:%s:%s:account:%s", resource.Type, region, resource.ID)
	}
}

func hasCycle(resourceID string, dependencyMap map[string][]string, visited, recStack map[string]bool) bool {
	visited[resourceID] = true
	recStack[resourceID] = true

	for _, dep := range dependencyMap[resourceID] {
		if !visited[dep] && hasCycle(dep, dependencyMap, visited, recStack) {
			return true
		} else if recStack[dep] {
			return true
		}
	}

	recStack[resourceID] = false
	return false
}

func topologicalSort(resources []CloudResource, dependencyMap map[string][]string) []string {
	var result []string
	visited := make(map[string]bool)

	var visit func(string)
	visit = func(resourceID string) {
		if visited[resourceID] {
			return
		}
		visited[resourceID] = true

		for _, dep := range dependencyMap[resourceID] {
			visit(dep)
		}

		result = append(result, resourceID)
	}

	for _, resource := range resources {
		visit(resource.ID)
	}

	return result
}

// Rollback Functions

func destroyNetworkingResources(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Destroying networking resources...")

	provisionedResources := ctx.GetData("provisionedResources").(map[string]CloudResource)

	destroyCount := 0
	for resourceID, resource := range provisionedResources {
		if resource.Type == "vpc" || resource.Type == "subnet" {
			ctx.Log(fmt.Sprintf("Destroying %s: %s", resource.Type, resource.Name))
			delete(provisionedResources, resourceID)
			destroyCount++
		}
	}

	ctx.SetData("provisionedResources", provisionedResources)

	message := fmt.Sprintf("Destroyed %d networking resources", destroyCount)
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}

func destroyComputeResources(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Destroying compute resources...")

	provisionedResources := ctx.GetData("provisionedResources").(map[string]CloudResource)

	destroyCount := 0
	for resourceID, resource := range provisionedResources {
		if resource.Type == "instance" {
			ctx.Log(fmt.Sprintf("Destroying %s: %s", resource.Type, resource.Name))
			delete(provisionedResources, resourceID)
			destroyCount++
		}
	}

	ctx.SetData("provisionedResources", provisionedResources)

	message := fmt.Sprintf("Destroyed %d compute resources", destroyCount)
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}

func destroyDatabaseResources(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Destroying database resources...")

	provisionedResources := ctx.GetData("provisionedResources").(map[string]CloudResource)

	destroyCount := 0
	for resourceID, resource := range provisionedResources {
		if resource.Type == "database" {
			ctx.Log(fmt.Sprintf("Destroying %s: %s", resource.Type, resource.Name))
			delete(provisionedResources, resourceID)
			destroyCount++
		}
	}

	ctx.SetData("provisionedResources", provisionedResources)

	message := fmt.Sprintf("Destroyed %d database resources", destroyCount)
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}

func destroyLoadBalancerResources(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Destroying load balancer resources...")

	provisionedResources := ctx.GetData("provisionedResources").(map[string]CloudResource)

	destroyCount := 0
	for resourceID, resource := range provisionedResources {
		if resource.Type == "loadbalancer" {
			ctx.Log(fmt.Sprintf("Destroying %s: %s", resource.Type, resource.Name))
			delete(provisionedResources, resourceID)
			destroyCount++
		}
	}

	ctx.SetData("provisionedResources", provisionedResources)

	message := fmt.Sprintf("Destroyed %d load balancer resources", destroyCount)
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}

func destroyProvisionedResources(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Destroying all provisioned resources...")

	provisionedResources := ctx.GetData("provisionedResources").(map[string]CloudResource)

	destroyCount := len(provisionedResources)
	for _, resource := range provisionedResources {
		ctx.Log(fmt.Sprintf("Destroying %s: %s", resource.Type, resource.Name))
	}

	// Clear all provisioned resources
	ctx.SetData("provisionedResources", make(map[string]CloudResource))

	message := fmt.Sprintf("Destroyed %d resources", destroyCount)
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}

func resetResourceConfiguration(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Resetting resource configurations...")

	job := ctx.GetData("job").(ProvisioningJob)
	job.Status = "configuration_reset"
	ctx.SetData("job", job)

	message := "Resource configurations reset to default state"
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}
