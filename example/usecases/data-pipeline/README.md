# Data Processing Pipeline

This example demonstrates a comprehensive data processing pipeline using GoSteps with the following features:

## Features Demonstrated

### 1. **Sequential Step Execution**
- Input validation
- Data reading and parsing
- Data processing with conditional branching
- Output validation
- Results saving

### 2. **Retry Logic with Different Strategies**
- Maximum retry attempts configuration
- Retry sleep intervals
- Retry on specific errors vs all errors
- Retry based on error patterns using regex

### 3. **Conditional Branching**
- Dynamic branch selection based on data size
- **Transform Pipeline**: For larger datasets (>3 records)
  - Data cleaning and validation
  - Data enrichment with additional fields
- **Aggregate Pipeline**: For smaller datasets
  - Data grouping by category
  - Metrics calculation

### 4. **Rollback Functionality**
- Automatic rollback on step failures
- Multiple rollback strategies:
  - Temporary file cleanup
  - Original data restoration
  - Output file cleanup

### 5. **Comprehensive Logging**
- Step-level logging with structured output
- Progress tracking and error reporting
- Log output to both console and file

### 6. **Context Data Management**
- Data sharing between steps
- Progress tracking
- Error counting and reporting

## Pipeline Flow

```
┌─────────────────┐    ┌──────────────┐    ┌─────────────────┐
│ Validate Input  │ -> │  Read Data   │ -> │ Process Data    │
└─────────────────┘    └──────────────┘    └─────────────────┘
                                                     │
                                          ┌─────────┴─────────┐
                                          │                   │
                                   ┌──────▼──────┐    ┌──────▼──────┐
                                   │ Transform   │    │ Aggregate   │
                                   │  Pipeline   │    │  Pipeline   │
                                   │             │    │             │
                                   │ • Clean     │    │ • Group     │
                                   │ • Enrich    │    │ • Metrics   │
                                   └─────────────┘    └─────────────┘
                                          │                   │
                                          └─────────┬─────────┘
                                                    │
                                          ┌─────────▼─────────┐
                                          │ Validate Output   │
                                          └─────────┬─────────┘
                                                    │
                                          ┌─────────▼─────────┐
                                          │  Save Results     │
                                          └───────────────────┘
```

## Running the Example

```bash
cd example/usecases/data-pipeline
go run main.go
```

## Sample Output

The pipeline processes sample data and demonstrates:
- Email validation and correction
- Data enrichment with tiers and regions
- Different processing paths based on data size
- Comprehensive logging of each step
- Error handling and retry mechanisms

## Key Learning Points

1. **Step Configuration**: Each step can have different retry policies and rollback functions
2. **Dynamic Branching**: The resolver function determines the execution path at runtime
3. **Error Handling**: Different types of errors can trigger different retry strategies
4. **Data Flow**: Context data flows seamlessly between steps and branches
5. **Rollback Safety**: Failed operations can be automatically reverted to maintain data integrity

This example showcases how GoSteps can be used to build robust, fault-tolerant data processing pipelines with complex business logic and error recovery mechanisms.