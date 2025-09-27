package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"time"

	gosteps "github.com/TanmoySG/go-steps"
	"github.com/rs/zerolog"
)

// MigrationJob represents a microservice data migration
type MigrationJob struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	SourceService   ServiceConfig          `json:"source_service"`
	TargetService   ServiceConfig          `json:"target_service"`
	MigrationPlan   MigrationPlan          `json:"migration_plan"`
	Status          string                 `json:"status"`
	StartTime       time.Time              `json:"start_time"`
	Progress        MigrationProgress      `json:"progress"`
	ValidationRules []ValidationRule       `json:"validation_rules"`
	Rollback        RollbackConfiguration  `json:"rollback"`
	Metadata        map[string]interface{} `json:"metadata"`
}

type ServiceConfig struct {
	Name           string            `json:"name"`
	Type           string            `json:"type"` // database, api, queue, cache
	ConnectionInfo map[string]string `json:"connection_info"`
	Schema         interface{}       `json:"schema"`
}

type MigrationPlan struct {
	BatchSize       int               `json:"batch_size"`
	TotalRecords    int               `json:"total_records"`
	MigrationMode   string            `json:"migration_mode"` // full, incremental, live
	DataMapping     map[string]string `json:"data_mapping"`
	TransformRules  []TransformRule   `json:"transform_rules"`
	DependencyCheck bool              `json:"dependency_check"`
}

type TransformRule struct {
	Field      string      `json:"field"`
	Operation  string      `json:"operation"` // rename, convert, calculate, default
	Parameters interface{} `json:"parameters"`
	Condition  string      `json:"condition,omitempty"`
}

type ValidationRule struct {
	Name       string      `json:"name"`
	Type       string      `json:"type"` // data_integrity, referential, business
	Parameters interface{} `json:"parameters"`
	Critical   bool        `json:"critical"`
}

type MigrationProgress struct {
	RecordsProcessed   int           `json:"records_processed"`
	RecordsSuccessful  int           `json:"records_successful"`
	RecordsFailed      int           `json:"records_failed"`
	BatchesCompleted   int           `json:"batches_completed"`
	ValidationsPassed  int           `json:"validations_passed"`
	ValidationsFailed  int           `json:"validations_failed"`
	CurrentPhase       string        `json:"current_phase"`
	EstimatedRemaining time.Duration `json:"estimated_remaining"`
}

type RollbackConfiguration struct {
	Enabled          bool                   `json:"enabled"`
	BackupLocation   string                 `json:"backup_location"`
	RollbackStrategy string                 `json:"rollback_strategy"` // snapshot, log_replay, selective
	RetentionPeriod  time.Duration          `json:"retention_period"`
	CheckpointData   map[string]interface{} `json:"checkpoint_data"`
}

func main() {
	// Setup logging
	runLogFile, _ := os.OpenFile(
		"migration.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	output := zerolog.MultiLevelWriter(os.Stdout, runLogFile)
	logger := gosteps.NewGoStepsLogger(output, &gosteps.LoggerOpts{StepLoggingEnabled: true})

	// Initialize context with sample migration job
	ctx := gosteps.NewGoStepsContext()
	ctx.Use(logger)

	sampleMigration := MigrationJob{
		ID:   "MIGRATION-2025-001",
		Name: "User Data Migration - Legacy to Modern",
		SourceService: ServiceConfig{
			Name: "legacy-user-service",
			Type: "database",
			ConnectionInfo: map[string]string{
				"host":     "legacy-db.internal",
				"database": "user_legacy",
				"schema":   "public",
			},
		},
		TargetService: ServiceConfig{
			Name: "modern-user-service",
			Type: "database",
			ConnectionInfo: map[string]string{
				"host":     "modern-db.cluster",
				"database": "users_v2",
				"schema":   "app",
			},
		},
		MigrationPlan: MigrationPlan{
			BatchSize:       1000,
			TotalRecords:    50000,
			MigrationMode:   "incremental",
			DependencyCheck: true,
			DataMapping: map[string]string{
				"user_id":    "id",
				"full_name":  "name",
				"email_addr": "email",
				"created_dt": "created_at",
			},
			TransformRules: []TransformRule{
				{
					Field:     "phone",
					Operation: "convert",
					Parameters: map[string]string{
						"format": "e164",
						"region": "US",
					},
				},
				{
					Field:      "status",
					Operation:  "rename",
					Parameters: "user_status",
				},
			},
		},
		ValidationRules: []ValidationRule{
			{
				Name:     "email_format",
				Type:     "data_integrity",
				Critical: true,
				Parameters: map[string]string{
					"pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
				},
			},
			{
				Name:     "unique_email",
				Type:     "referential",
				Critical: true,
			},
		},
		Status:    "pending",
		StartTime: time.Now(),
		Progress:  MigrationProgress{CurrentPhase: "initialization"},
		Rollback: RollbackConfiguration{
			Enabled:          true,
			BackupLocation:   "/backups/migration-001",
			RollbackStrategy: "snapshot",
			RetentionPeriod:  72 * time.Hour,
			CheckpointData:   make(map[string]interface{}),
		},
		Metadata: map[string]interface{}{
			"initiated_by": "migration-service",
			"priority":     "high",
			"environment":  "production",
		},
	}

	ctx.WithData(map[string]interface{}{
		"migration":     sampleMigration,
		"backupCreated": false,
		"connections":   make(map[string]interface{}),
		"checkpoints":   []string{},
		"errorLog":      []string{},
	})

	// Define the migration workflow
	steps := gosteps.Steps{
		{
			Name:     "validateMigrationPlan",
			Function: validateMigrationPlanStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 2,
				RetrySleep:     1 * time.Second,
			},
		},
		{
			Name:     "createBackup",
			Function: createBackupStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 3,
				RetrySleep:     5 * time.Second,
			},
			RollbackFunction: cleanupBackup,
		},
		{
			Name:     "establishConnections",
			Function: establishConnectionsStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 3,
				ErrorPatternsToRetry: []regexp.Regexp{
					*regexp.MustCompile("connection.*timeout"),
					*regexp.MustCompile("database.*unavailable"),
				},
				RetrySleep: 3 * time.Second,
			},
			RollbackFunction: closeConnections,
		},
		{
			Name:     "validateDependencies",
			Function: validateDependenciesStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 2,
			},
		},
		{
			Name:     "executeMigration",
			Function: executeMigrationStep,
			Branches: &gosteps.Branches{
				Resolver: migrationModeResolver,
				Branches: []gosteps.Branch{
					{
						BranchName: "fullMigration",
						Steps: gosteps.Steps{
							{
								Name:     "performFullDataCopy",
								Function: performFullDataCopyStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 2,
									RetrySleep:     10 * time.Second,
								},
								RollbackFunction: rollbackMigration,
							},
							{
								Name:     "validateFullMigration",
								Function: validateFullMigrationStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 3,
									RetrySleep:     5 * time.Second,
								},
							},
						},
					},
					{
						BranchName: "incrementalMigration",
						Steps: gosteps.Steps{
							{
								Name:             "createCheckpoint",
								Function:         createCheckpointStep,
								RollbackFunction: cleanupCheckpoints,
							},
							{
								Name:     "performIncrementalCopy",
								Function: performIncrementalCopyStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 3,
									RetrySleep:     5 * time.Second,
								},
								RollbackFunction: rollbackMigration,
							},
							{
								Name:     "validateIncrementalData",
								Function: validateIncrementalDataStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 2,
								},
							},
							{
								Name:     "updateCheckpoint",
								Function: updateCheckpointStep,
							},
						},
					},
					{
						BranchName: "liveMigration",
						Steps: gosteps.Steps{
							{
								Name:             "setupChangeCapture",
								Function:         setupChangeCaptureStep,
								RollbackFunction: cleanupChangeCapture,
							},
							{
								Name:     "performInitialSync",
								Function: performInitialSyncStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 2,
									RetrySleep:     10 * time.Second,
								},
								RollbackFunction: rollbackMigration,
							},
							{
								Name:     "applyContinuousChanges",
								Function: applyContinuousChangesStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 5,
									RetrySleep:     2 * time.Second,
								},
							},
						},
					},
				},
			},
		},
		{
			Name:     "performDataValidation",
			Function: performDataValidationStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 3,
				RetrySleep:     2 * time.Second,
			},
		},
		{
			Name:     "generateMigrationReport",
			Function: generateMigrationReportStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 2,
			},
		},
		{
			Name:     "cleanupResources",
			Function: cleanupResourcesStep,
		},
	}

	// Execute the migration workflow
	fmt.Println("🔄 Starting Microservice Data Migration...")
	stepsProcessor := gosteps.NewStepsProcessor(steps)
	stepsProcessor.Execute(ctx)

	// Print final migration status
	finalMigration := ctx.GetData("migration").(MigrationJob)
	fmt.Printf("\n📊 Migration Complete!\n")
	fmt.Printf("Migration ID: %s\n", finalMigration.ID)
	fmt.Printf("Status: %s\n", finalMigration.Status)
	fmt.Printf("Records Processed: %d/%d\n", finalMigration.Progress.RecordsProcessed, finalMigration.MigrationPlan.TotalRecords)
	fmt.Printf("Success Rate: %.2f%%\n", float64(finalMigration.Progress.RecordsSuccessful)/float64(finalMigration.Progress.RecordsProcessed)*100)
	fmt.Printf("Duration: %v\n", time.Since(finalMigration.StartTime).Round(time.Second))
}

// Step Functions

func validateMigrationPlanStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Validating migration plan configuration...")

	migration := ctx.GetData("migration").(MigrationJob)

	// Validate source and target services
	if migration.SourceService.Name == "" || migration.TargetService.Name == "" {
		return gosteps.MarkStateFailed().WithMessage("Source and target services must be specified")
	}

	// Validate migration plan
	if migration.MigrationPlan.BatchSize <= 0 {
		return gosteps.MarkStateFailed().WithMessage("Batch size must be greater than zero")
	}

	if migration.MigrationPlan.TotalRecords <= 0 {
		return gosteps.MarkStateFailed().WithMessage("Total records must be specified")
	}

	// Validate data mappings
	if len(migration.MigrationPlan.DataMapping) == 0 {
		return gosteps.MarkStateFailed().WithMessage("Data mapping configuration is required")
	}

	migration.Status = "plan_validated"
	migration.Progress.CurrentPhase = "plan_validation_complete"
	ctx.SetData("migration", migration)

	return gosteps.MarkStateComplete().WithMessage("Migration plan validation successful")
}

func createBackupStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Creating data backup before migration...")

	migration := ctx.GetData("migration").(MigrationJob)

	if !migration.Rollback.Enabled {
		return gosteps.MarkStateSkipped().WithMessage("Backup disabled in configuration")
	}

	// Simulate backup creation
	time.Sleep(200 * time.Millisecond)

	// Simulate potential backup failures
	if rand.Float32() < 0.1 {
		return gosteps.MarkStateError().WithError(fmt.Errorf("backup storage unavailable"))
	}

	backupID := fmt.Sprintf("backup-%d", time.Now().UnixNano())
	migration.Rollback.CheckpointData["backup_id"] = backupID
	migration.Status = "backup_created"
	migration.Progress.CurrentPhase = "backup_complete"

	ctx.SetData("migration", migration)
	ctx.SetData("backupCreated", true)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Backup created successfully: %s", backupID)).
		WithData(map[string]interface{}{
			"backupID": backupID,
		})
}

func establishConnectionsStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Establishing database connections...")

	migration := ctx.GetData("migration").(MigrationJob)

	// Simulate connection establishment with potential failures
	if rand.Float32() < 0.15 {
		return gosteps.MarkStateError().WithError(fmt.Errorf("connection timeout to %s", migration.SourceService.Name))
	}

	// Create mock connections
	connections := map[string]interface{}{
		"source": map[string]string{
			"connection_id": fmt.Sprintf("conn-src-%d", time.Now().UnixNano()),
			"status":        "connected",
		},
		"target": map[string]string{
			"connection_id": fmt.Sprintf("conn-tgt-%d", time.Now().UnixNano()),
			"status":        "connected",
		},
	}

	migration.Status = "connections_established"
	migration.Progress.CurrentPhase = "connections_ready"
	ctx.SetData("migration", migration)
	ctx.SetData("connections", connections)

	return gosteps.MarkStateComplete().WithMessage("Database connections established successfully")
}

func validateDependenciesStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Validating migration dependencies...")

	migration := ctx.GetData("migration").(MigrationJob)

	if !migration.MigrationPlan.DependencyCheck {
		return gosteps.MarkStateSkipped().WithMessage("Dependency check disabled")
	}

	// Simulate dependency validation
	dependencies := []string{"foreign_key_constraints", "index_dependencies", "trigger_dependencies"}
	for _, dep := range dependencies {
		ctx.Log(fmt.Sprintf("Checking dependency: %s", dep))
		time.Sleep(10 * time.Millisecond)
	}

	migration.Status = "dependencies_validated"
	migration.Progress.CurrentPhase = "dependency_check_complete"
	ctx.SetData("migration", migration)

	return gosteps.MarkStateComplete().WithMessage("All dependencies validated successfully")
}

func executeMigrationStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Determining migration execution strategy...")

	migration := ctx.GetData("migration").(MigrationJob)
	migration.Status = "executing"
	migration.Progress.CurrentPhase = "migration_execution"
	ctx.SetData("migration", migration)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Executing %s migration strategy", migration.MigrationPlan.MigrationMode))
}

func migrationModeResolver(ctx gosteps.GoStepsCtx) gosteps.BranchName {
	migration := ctx.GetData("migration").(MigrationJob)

	switch migration.MigrationPlan.MigrationMode {
	case "full":
		ctx.Log("Using full migration strategy")
		return gosteps.BranchName("fullMigration")
	case "incremental":
		ctx.Log("Using incremental migration strategy")
		return gosteps.BranchName("incrementalMigration")
	case "live":
		ctx.Log("Using live migration strategy")
		return gosteps.BranchName("liveMigration")
	default:
		ctx.Log("Defaulting to incremental migration")
		return gosteps.BranchName("incrementalMigration")
	}
}

func performFullDataCopyStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Performing full data copy...")

	migration := ctx.GetData("migration").(MigrationJob)

	// Simulate full data copy with progress updates
	totalRecords := migration.MigrationPlan.TotalRecords
	batchSize := migration.MigrationPlan.BatchSize
	batches := (totalRecords + batchSize - 1) / batchSize

	for i := 0; i < batches; i++ {
		recordsInBatch := batchSize
		if i == batches-1 {
			recordsInBatch = totalRecords - (i * batchSize)
		}

		// Simulate batch processing
		time.Sleep(20 * time.Millisecond)

		// Simulate potential batch failures
		if rand.Float32() < 0.05 {
			return gosteps.MarkStateError().WithError(fmt.Errorf("batch %d failed due to data constraint violation", i+1))
		}

		migration.Progress.RecordsProcessed += recordsInBatch
		migration.Progress.RecordsSuccessful += recordsInBatch
		migration.Progress.BatchesCompleted++

		ctx.Log(fmt.Sprintf("Batch %d completed: %d records", i+1, recordsInBatch))
	}

	migration.Status = "data_copied"
	migration.Progress.CurrentPhase = "full_copy_complete"
	ctx.SetData("migration", migration)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Full data copy completed: %d records in %d batches",
			migration.Progress.RecordsProcessed, migration.Progress.BatchesCompleted))
}

func validateFullMigrationStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Validating full migration results...")

	migration := ctx.GetData("migration").(MigrationJob)

	// Simulate validation process
	time.Sleep(100 * time.Millisecond)

	// Run validation rules
	validationsPassed := 0
	validationsFailed := 0

	for _, rule := range migration.ValidationRules {
		ctx.Log(fmt.Sprintf("Running validation: %s", rule.Name))

		// Simulate validation with potential failures
		if rand.Float32() < 0.1 && rule.Critical {
			validationsFailed++
			if rule.Critical {
				return gosteps.MarkStateFailed().
					WithMessage(fmt.Sprintf("Critical validation failed: %s", rule.Name))
			}
		} else {
			validationsPassed++
		}
	}

	migration.Progress.ValidationsPassed = validationsPassed
	migration.Progress.ValidationsFailed = validationsFailed
	migration.Status = "validation_complete"
	migration.Progress.CurrentPhase = "full_migration_validated"
	ctx.SetData("migration", migration)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Validation complete: %d passed, %d failed",
			validationsPassed, validationsFailed))
}

func createCheckpointStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Creating migration checkpoint...")

	migration := ctx.GetData("migration").(MigrationJob)

	checkpointID := fmt.Sprintf("checkpoint-%d", time.Now().UnixNano())
	migration.Rollback.CheckpointData["current_checkpoint"] = checkpointID

	checkpoints := ctx.GetData("checkpoints").([]string)
	checkpoints = append(checkpoints, checkpointID)
	ctx.SetData("checkpoints", checkpoints)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Checkpoint created: %s", checkpointID))
}

func performIncrementalCopyStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Performing incremental data copy...")

	migration := ctx.GetData("migration").(MigrationJob)

	// Simulate incremental copy with smaller batch sizes
	incrementalRecords := migration.MigrationPlan.TotalRecords / 2 // Simulate partial data
	batchSize := migration.MigrationPlan.BatchSize
	batches := (incrementalRecords + batchSize - 1) / batchSize

	for i := 0; i < batches; i++ {
		recordsInBatch := batchSize
		if i == batches-1 {
			recordsInBatch = incrementalRecords - (i * batchSize)
		}

		time.Sleep(15 * time.Millisecond)

		migration.Progress.RecordsProcessed += recordsInBatch
		migration.Progress.RecordsSuccessful += recordsInBatch
		migration.Progress.BatchesCompleted++
	}

	migration.Status = "incremental_copy_complete"
	migration.Progress.CurrentPhase = "incremental_data_copied"
	ctx.SetData("migration", migration)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Incremental copy completed: %d records processed",
			migration.Progress.RecordsProcessed))
}

func validateIncrementalDataStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Validating incremental migration data...")

	migration := ctx.GetData("migration").(MigrationJob)

	// Simplified validation for incremental data
	validationsPassed := len(migration.ValidationRules)
	migration.Progress.ValidationsPassed = validationsPassed
	migration.Status = "incremental_validated"
	migration.Progress.CurrentPhase = "incremental_validation_complete"
	ctx.SetData("migration", migration)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Incremental data validation complete: %d validations passed",
			validationsPassed))
}

func updateCheckpointStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Updating migration checkpoint...")

	migration := ctx.GetData("migration").(MigrationJob)

	migration.Rollback.CheckpointData["records_processed"] = migration.Progress.RecordsProcessed
	migration.Rollback.CheckpointData["last_update"] = time.Now().Format(time.RFC3339)

	ctx.SetData("migration", migration)

	return gosteps.MarkStateComplete().WithMessage("Checkpoint updated with current progress")
}

func setupChangeCaptureStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Setting up change data capture...")

	changeCaptureID := fmt.Sprintf("cdc-%d", time.Now().UnixNano())

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Change data capture setup complete: %s", changeCaptureID)).
		WithData(map[string]interface{}{
			"changeCaptureID": changeCaptureID,
		})
}

func performInitialSyncStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Performing initial data synchronization...")

	migration := ctx.GetData("migration").(MigrationJob)

	// Simulate initial sync
	syncRecords := migration.MigrationPlan.TotalRecords * 3 / 4 // 75% of data
	migration.Progress.RecordsProcessed = syncRecords
	migration.Progress.RecordsSuccessful = syncRecords
	migration.Status = "initial_sync_complete"
	migration.Progress.CurrentPhase = "live_initial_sync_complete"

	ctx.SetData("migration", migration)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Initial sync completed: %d records synchronized", syncRecords))
}

func applyContinuousChangesStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Applying continuous changes...")

	migration := ctx.GetData("migration").(MigrationJob)

	// Simulate applying remaining changes
	remainingRecords := migration.MigrationPlan.TotalRecords - migration.Progress.RecordsProcessed

	if remainingRecords > 0 {
		migration.Progress.RecordsProcessed += remainingRecords
		migration.Progress.RecordsSuccessful += remainingRecords
	}

	migration.Status = "live_migration_complete"
	migration.Progress.CurrentPhase = "continuous_changes_applied"
	ctx.SetData("migration", migration)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Continuous changes applied: %d additional records", remainingRecords))
}

func performDataValidationStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Performing comprehensive data validation...")

	migration := ctx.GetData("migration").(MigrationJob)

	// Comprehensive validation after migration
	time.Sleep(150 * time.Millisecond)

	validationsPassed := len(migration.ValidationRules)
	validationsFailed := 0

	// Simulate some validation failures
	if rand.Float32() < 0.2 {
		validationsFailed = 1
		validationsPassed -= 1
	}

	migration.Progress.ValidationsPassed = validationsPassed
	migration.Progress.ValidationsFailed = validationsFailed
	migration.Status = "data_validated"
	migration.Progress.CurrentPhase = "comprehensive_validation_complete"
	ctx.SetData("migration", migration)

	if validationsFailed > 0 {
		return gosteps.MarkStateError().WithError(fmt.Errorf("data validation failures detected: %d", validationsFailed))
	}

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Data validation successful: %d validations passed", validationsPassed))
}

func generateMigrationReportStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Generating migration report...")

	migration := ctx.GetData("migration").(MigrationJob)

	// Generate comprehensive migration report
	report := map[string]interface{}{
		"migration_id":       migration.ID,
		"total_records":      migration.MigrationPlan.TotalRecords,
		"records_processed":  migration.Progress.RecordsProcessed,
		"records_successful": migration.Progress.RecordsSuccessful,
		"records_failed":     migration.Progress.RecordsFailed,
		"batches_completed":  migration.Progress.BatchesCompleted,
		"validations_passed": migration.Progress.ValidationsPassed,
		"validations_failed": migration.Progress.ValidationsFailed,
		"duration":           time.Since(migration.StartTime).String(),
		"success_rate":       float64(migration.Progress.RecordsSuccessful) / float64(migration.Progress.RecordsProcessed) * 100,
	}

	reportJSON, _ := json.MarshalIndent(report, "", "  ")
	ctx.Log(fmt.Sprintf("Migration Report:\n%s", string(reportJSON)))

	migration.Status = "report_generated"
	migration.Progress.CurrentPhase = "migration_complete"
	ctx.SetData("migration", migration)

	return gosteps.MarkStateComplete().
		WithMessage("Migration report generated successfully").
		WithData(map[string]interface{}{
			"report": report,
		})
}

func cleanupResourcesStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Cleaning up migration resources...")

	migration := ctx.GetData("migration").(MigrationJob)

	// Cleanup temporary resources
	cleanupItems := []string{
		"temporary_tables",
		"staging_data",
		"migration_locks",
		"temp_connections",
	}

	for _, item := range cleanupItems {
		ctx.Log(fmt.Sprintf("Cleaning up: %s", item))
		time.Sleep(5 * time.Millisecond)
	}

	migration.Status = "completed"
	migration.Progress.CurrentPhase = "cleanup_complete"
	ctx.SetData("migration", migration)

	return gosteps.MarkStateComplete().WithMessage("Resource cleanup completed successfully")
}

// Rollback Functions

func cleanupBackup(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Cleaning up backup files...")

	backupCreated := ctx.GetData("backupCreated").(bool)
	if !backupCreated {
		message := "No backup to cleanup"
		return gosteps.RollbackResult{
			RollbackState:   gosteps.RollbackStateSuccess,
			RollbackMessage: &message,
		}
	}

	// Simulate backup cleanup
	ctx.SetData("backupCreated", false)

	message := "Backup files cleaned up successfully"
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}

func closeConnections(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Closing database connections...")

	connections := ctx.GetData("connections").(map[string]interface{})

	for connType := range connections {
		ctx.Log(fmt.Sprintf("Closing %s connection", connType))
	}

	// Clear connections
	ctx.SetData("connections", make(map[string]interface{}))

	message := "Database connections closed successfully"
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}

func rollbackMigration(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Rolling back migration changes...")

	migration := ctx.GetData("migration").(MigrationJob)

	if migration.Rollback.RollbackStrategy == "snapshot" {
		backupID := migration.Rollback.CheckpointData["backup_id"]
		if backupID != nil {
			ctx.Log(fmt.Sprintf("Restoring from backup: %s", backupID))

			migration.Status = "rolled_back"
			migration.Progress.CurrentPhase = "rollback_complete"
			ctx.SetData("migration", migration)

			message := fmt.Sprintf("Migration rolled back from backup: %s", backupID)
			return gosteps.RollbackResult{
				RollbackState:   gosteps.RollbackStateSuccess,
				RollbackMessage: &message,
			}
		}
	}

	message := "Rollback completed - no backup available"
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateFailed,
		RollbackMessage: &message,
		RollbackError:   fmt.Errorf("no backup available for rollback"),
	}
}

func cleanupCheckpoints(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Cleaning up migration checkpoints...")

	checkpoints := ctx.GetData("checkpoints").([]string)

	for _, checkpoint := range checkpoints {
		ctx.Log(fmt.Sprintf("Removing checkpoint: %s", checkpoint))
	}

	ctx.SetData("checkpoints", []string{})

	message := fmt.Sprintf("Cleaned up %d checkpoints", len(checkpoints))
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}

func cleanupChangeCapture(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Cleaning up change data capture...")

	changeCaptureID := ctx.GetData("changeCaptureID")
	if changeCaptureID != nil {
		ctx.Log(fmt.Sprintf("Removing CDC setup: %s", changeCaptureID))
		ctx.SetData("changeCaptureID", nil)
	}

	message := "Change data capture cleanup completed"
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}
