package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	gosteps "github.com/TanmoySG/go-steps"
	"github.com/rs/zerolog"
)

// DataRecord represents a record in our data processing pipeline
type DataRecord struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
}

func main() {
	// Setup logging
	runLogFile, _ := os.OpenFile(
		"data-pipeline.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	output := zerolog.MultiLevelWriter(os.Stdout, runLogFile)
	logger := gosteps.NewGoStepsLogger(output, &gosteps.LoggerOpts{StepLoggingEnabled: true})

	// Initialize context with sample data
	ctx := gosteps.NewGoStepsContext()
	ctx.Use(logger)
	ctx.WithData(map[string]interface{}{
		"inputFile":    "sample_data.csv",
		"outputFile":   "processed_data.json",
		"batchSize":    100,
		"processCount": 0,
		"errorCount":   0,
	})

	// Define the data processing pipeline
	steps := gosteps.Steps{
		{
			Name:     "validateInput",
			Function: validateInputStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 3,
				RetrySleep:     1 * time.Second,
			},
			RollbackFunction: cleanupTempFiles,
		},
		{
			Name:     "readData",
			Function: readDataStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 2,
				RetrySleep:     2 * time.Second,
			},
			RollbackFunction: cleanupTempFiles,
		},
		{
			Name:     "processData",
			Function: processDataStep,
			Branches: &gosteps.Branches{
				Resolver: dataProcessingResolver,
				Branches: []gosteps.Branch{
					{
						BranchName: "transformData",
						Steps: gosteps.Steps{
							{
								Name:     "cleanData",
								Function: cleanDataStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 3,
									RetryAllErrors: true,
									RetrySleep:     1 * time.Second,
								},
								RollbackFunction: restoreOriginalData,
							},
							{
								Name:             "enrichData",
								Function:         enrichDataStep,
								RollbackFunction: restoreOriginalData,
							},
						},
					},
					{
						BranchName: "aggregateData",
						Steps: gosteps.Steps{
							{
								Name:             "groupData",
								Function:         groupDataStep,
								RollbackFunction: restoreOriginalData,
							},
							{
								Name:             "calculateMetrics",
								Function:         calculateMetricsStep,
								RollbackFunction: restoreOriginalData,
							},
						},
					},
				},
			},
		},
		{
			Name:     "validateOutput",
			Function: validateOutputStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 2,
				ErrorPatternsToRetry: []regexp.Regexp{
					*regexp.MustCompile("validation.*error"),
				},
			},
		},
		{
			Name:     "saveResults",
			Function: saveResultsStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 3,
				RetrySleep:     2 * time.Second,
			},
			RollbackFunction: cleanupOutputFiles,
		},
	}

	// Execute the pipeline
	fmt.Println("🚀 Starting Data Processing Pipeline...")
	stepsProcessor := gosteps.NewStepsProcessor(steps)
	stepsProcessor.Execute(ctx)

	fmt.Println("\n📊 Pipeline Execution Summary:")
	fmt.Printf("Records Processed: %v\n", ctx.GetData("processCount"))
	fmt.Printf("Errors Encountered: %v\n", ctx.GetData("errorCount"))
}

// Step Functions

func validateInputStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Validating input parameters...")

	inputFile := ctx.GetData("inputFile").(string)
	if inputFile == "" {
		return gosteps.MarkStateError().WithError(fmt.Errorf("input file not specified"))
	}

	// Check if input file exists (simulate)
	if inputFile == "missing_file.csv" {
		return gosteps.MarkStateError().WithError(fmt.Errorf("input file does not exist: %s", inputFile))
	}

	batchSize := ctx.GetData("batchSize").(int)
	if batchSize <= 0 {
		return gosteps.MarkStateFailed().WithMessage("Invalid batch size")
	}

	return gosteps.MarkStateComplete().WithMessage("Input validation successful")
}

func readDataStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Reading data from input source...")

	// Simulate reading CSV data
	sampleData := []DataRecord{
		{ID: 1, Name: "John Doe", Email: "john@example.com", Amount: 100.50, Category: "premium"},
		{ID: 2, Name: "Jane Smith", Email: "jane@example.com", Amount: 75.25, Category: "standard"},
		{ID: 3, Name: "Bob Johnson", Email: "bob@example.com", Amount: 200.00, Category: "premium"},
		{ID: 4, Name: "Alice Brown", Email: "invalid-email", Amount: 50.00, Category: "basic"},
		{ID: 5, Name: "", Email: "empty@example.com", Amount: 150.00, Category: "standard"},
	}

	return gosteps.MarkStateComplete().
		WithData(map[string]interface{}{
			"rawData":     sampleData,
			"recordCount": len(sampleData),
		}).
		WithMessage(fmt.Sprintf("Successfully read %d records", len(sampleData)))
}

func processDataStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Determining data processing strategy...")

	recordCount := ctx.GetData("recordCount").(int)

	// Store processing metadata
	ctx.SetData("processingStartTime", time.Now())

	return gosteps.MarkStateComplete().WithMessage(fmt.Sprintf("Processing %d records", recordCount))
}

func dataProcessingResolver(ctx gosteps.GoStepsCtx) gosteps.BranchName {
	recordCount := ctx.GetData("recordCount").(int)

	// Choose processing strategy based on data size
	if recordCount > 3 {
		ctx.Log("Large dataset detected - using transformation pipeline")
		return gosteps.BranchName("transformData")
	} else {
		ctx.Log("Small dataset detected - using aggregation pipeline")
		return gosteps.BranchName("aggregateData")
	}
}

func cleanDataStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Cleaning and validating data...")

	rawData := ctx.GetData("rawData").([]DataRecord)
	var cleanedData []DataRecord
	errorCount := 0

	for _, record := range rawData {
		// Validate email format
		if !isValidEmail(record.Email) {
			ctx.Log(fmt.Sprintf("Invalid email for record ID %d: %s", record.ID, record.Email))
			record.Email = "invalid@placeholder.com"
			errorCount++
		}

		// Validate name
		if record.Name == "" {
			ctx.Log(fmt.Sprintf("Empty name for record ID %d", record.ID))
			record.Name = "Unknown"
			errorCount++
		}

		cleanedData = append(cleanedData, record)
	}

	// Simulate occasional cleaning failure
	if errorCount > 2 {
		return gosteps.MarkStatePending().WithMessage("Too many errors, retrying cleaning process")
	}

	return gosteps.MarkStateComplete().
		WithData(map[string]interface{}{
			"cleanedData": cleanedData,
			"errorCount":  errorCount,
		}).
		WithMessage(fmt.Sprintf("Data cleaned successfully, %d errors corrected", errorCount))
}

func enrichDataStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Enriching data with additional information...")

	cleanedData := ctx.GetData("cleanedData").([]DataRecord)
	var enrichedData []map[string]interface{}

	for _, record := range cleanedData {
		enriched := map[string]interface{}{
			"id":        record.ID,
			"name":      record.Name,
			"email":     record.Email,
			"amount":    record.Amount,
			"category":  record.Category,
			"tier":      calculateTier(record.Amount),
			"region":    "US", // Simulated enrichment
			"timestamp": time.Now().Format(time.RFC3339),
		}
		enrichedData = append(enrichedData, enriched)
	}

	processCount := len(enrichedData)
	ctx.SetData("processCount", processCount)

	return gosteps.MarkStateComplete().
		WithData(map[string]interface{}{
			"enrichedData": enrichedData,
		}).
		WithMessage(fmt.Sprintf("Data enriched successfully for %d records", processCount))
}

func groupDataStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Grouping data by category...")

	rawData := ctx.GetData("rawData").([]DataRecord)
	groupedData := make(map[string][]DataRecord)

	for _, record := range rawData {
		groupedData[record.Category] = append(groupedData[record.Category], record)
	}

	processCount := len(rawData)
	ctx.SetData("processCount", processCount)

	return gosteps.MarkStateComplete().
		WithData(map[string]interface{}{
			"groupedData": groupedData,
		}).
		WithMessage(fmt.Sprintf("Data grouped into %d categories", len(groupedData)))
}

func calculateMetricsStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Calculating metrics for grouped data...")

	groupedData := ctx.GetData("groupedData").(map[string][]DataRecord)
	metrics := make(map[string]map[string]interface{})

	for category, records := range groupedData {
		totalAmount := 0.0
		count := len(records)

		for _, record := range records {
			totalAmount += record.Amount
		}

		avgAmount := totalAmount / float64(count)

		metrics[category] = map[string]interface{}{
			"count":    count,
			"total":    totalAmount,
			"average":  avgAmount,
			"category": category,
		}
	}

	return gosteps.MarkStateComplete().
		WithData(map[string]interface{}{
			"metricsData": metrics,
		}).
		WithMessage(fmt.Sprintf("Metrics calculated for %d categories", len(metrics)))
}

func validateOutputStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Validating processed output...")

	// Check if we have enriched data or metrics data
	if enrichedData := ctx.GetData("enrichedData"); enrichedData != nil {
		data := enrichedData.([]map[string]interface{})
		if len(data) == 0 {
			return gosteps.MarkStateError().WithError(fmt.Errorf("validation error: no enriched data found"))
		}

		// Validate required fields
		for i, record := range data {
			if record["tier"] == nil {
				return gosteps.MarkStateError().WithError(fmt.Errorf("validation error: missing tier for record %d", i))
			}
		}
	} else if metricsData := ctx.GetData("metricsData"); metricsData != nil {
		data := metricsData.(map[string]map[string]interface{})
		if len(data) == 0 {
			return gosteps.MarkStateError().WithError(fmt.Errorf("validation error: no metrics data found"))
		}
	} else {
		return gosteps.MarkStateFailed().WithMessage("No processed data found for validation")
	}

	return gosteps.MarkStateComplete().WithMessage("Output validation successful")
}

func saveResultsStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Saving processed results...")

	outputFile := ctx.GetData("outputFile").(string)

	var dataToSave interface{}
	var dataType string

	if enrichedData := ctx.GetData("enrichedData"); enrichedData != nil {
		dataToSave = enrichedData
		dataType = "enriched"
	} else if metricsData := ctx.GetData("metricsData"); metricsData != nil {
		dataToSave = metricsData
		dataType = "metrics"
	} else {
		return gosteps.MarkStateFailed().WithMessage("No data to save")
	}

	// Simulate saving to file
	jsonData, err := json.MarshalIndent(dataToSave, "", "  ")
	if err != nil {
		return gosteps.MarkStateError().WithError(fmt.Errorf("failed to marshal data: %v", err))
	}

	// Simulate file writing (in real scenario, would write to actual file)
	ctx.Log(fmt.Sprintf("Would save %d bytes of %s data to %s", len(jsonData), dataType, outputFile))

	return gosteps.MarkStateComplete().
		WithData(map[string]interface{}{
			"outputSize": len(jsonData),
			"dataType":   dataType,
		}).
		WithMessage(fmt.Sprintf("Results saved successfully to %s", outputFile))
}

// Rollback Functions

func cleanupTempFiles(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Cleaning up temporary files...")

	// Simulate cleanup
	tempFiles := []string{"temp_data.tmp", "processing.lock"}
	for _, file := range tempFiles {
		ctx.Log(fmt.Sprintf("Removing temporary file: %s", file))
	}

	message := "Temporary files cleaned up successfully"
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}

func restoreOriginalData(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Restoring original data state...")

	// Simulate data restoration
	if rawData := ctx.GetData("rawData"); rawData != nil {
		ctx.SetData("processedData", nil)
		ctx.SetData("cleanedData", nil)

		message := "Original data state restored"
		return gosteps.RollbackResult{
			RollbackState:   gosteps.RollbackStateSuccess,
			RollbackMessage: &message,
		}
	}

	message := "Failed to restore original data - backup not found"
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateFailed,
		RollbackMessage: &message,
		RollbackError:   fmt.Errorf("backup data not available"),
	}
}

func cleanupOutputFiles(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Cleaning up output files...")

	outputFile := ctx.GetData("outputFile").(string)
	ctx.Log(fmt.Sprintf("Removing output file: %s", outputFile))

	message := "Output files cleaned up successfully"
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}

// Helper Functions

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	return emailRegex.MatchString(strings.ToLower(email))
}

func calculateTier(amount float64) string {
	if amount >= 150.0 {
		return "gold"
	} else if amount >= 100.0 {
		return "silver"
	} else {
		return "bronze"
	}
}
