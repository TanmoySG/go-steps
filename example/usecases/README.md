# GoSteps Usecase Applications

This directory contains comprehensive real-world applications demonstrating the power and flexibility of the GoSteps library. Each example showcases different combinations of GoSteps features in production-ready scenarios.

## Overview

GoSteps is a Go library that helps run functions as steps in a sequential chain with advanced features like conditional branching, retry logic, rollback mechanisms, and comprehensive logging. These examples demonstrate how GoSteps can be applied to solve complex workflow orchestration challenges.

## Featured Applications

### 1. 📊 [Data Processing Pipeline](./data-pipeline/)

**Scenario**: ETL pipeline for processing large datasets with data transformation and quality validation.

**Key Features**:
- **Sequential Processing**: Input validation → Data reading → Processing → Validation → Output
- **Conditional Branching**: Transform vs Aggregate pipelines based on data characteristics
- **Retry Logic**: Network timeouts, processing errors, validation failures
- **Rollback Mechanisms**: Restore original data, cleanup temporary files
- **Data Flow**: Context data sharing between transformation steps

**Business Value**: Reliable data processing with error recovery and data integrity guarantees.

### 2. 🛒 [E-commerce Order Processing](./ecommerce-order/)

**Scenario**: Complete order fulfillment workflow with payment processing, inventory management, and shipping.

**Key Features**:
- **Complex Workflow**: Order validation → Inventory check → Payment → Fulfillment → Confirmation
- **Multiple Fulfillment Paths**: Standard, Express, and Backorder processing
- **Advanced Retry Strategies**: Payment gateway timeouts, notification service failures
- **Sophisticated Rollbacks**: Payment refunds, inventory release, shipment cancellation
- **Real-time Inventory**: Stock tracking with reservation management

**Business Value**: Robust order processing ensuring customer satisfaction and business continuity.

### 3. 🚀 [CI/CD Pipeline Orchestrator](./cicd-pipeline/)

**Scenario**: Automated software delivery pipeline with testing, security scanning, and deployment.

**Key Features**:
- **Multi-Stage Pipeline**: Build → Test → Scan → Deploy → Notify
- **Dynamic Test Selection**: Unit, Integration, or Full test suites
- **Deployment Strategies**: Staging, Production, and Canary deployments
- **Comprehensive Retries**: Git operations, test execution, deployment verification
- **Automated Rollbacks**: Workspace cleanup, artifact removal, deployment reversion

**Business Value**: Reliable software delivery with quality gates and risk mitigation.

### 4. 🔄 [Microservice Migration Tool](./microservice-migration/)

**Scenario**: Data migration between microservices with validation, transformation, and zero-downtime support.

**Key Features**:
- **Multiple Migration Modes**: Full, Incremental, and Live migration strategies
- **Data Transformation**: Field mapping, format conversion, business rule application
- **Comprehensive Validation**: Data integrity, referential constraints, business rules
- **Advanced Rollbacks**: Snapshot restoration, checkpoint recovery, change log replay
- **Progress Tracking**: Real-time migration monitoring with detailed reporting

**Business Value**: Safe data migrations with minimal downtime and data integrity assurance.

### 5. ☁️ [Cloud Resource Provisioner](./cloud-provisioning/)

**Scenario**: Infrastructure as Code provisioning with dependency management and multi-cloud support.

**Key Features**:
- **Dependency Management**: Automatic resource ordering, circular dependency detection
- **Multiple Strategies**: Network-first (stability) vs Parallel (speed) provisioning
- **Resource Support**: VPC, Compute, Database, Load Balancer resources
- **Intelligent Rollbacks**: Dependency-aware resource cleanup
- **Production Features**: Authentication, validation, monitoring, reporting

**Business Value**: Reliable infrastructure provisioning with operational excellence and cost optimization.

## GoSteps Features Demonstrated

### Core Workflow Management
- **Sequential Execution**: Ordered step processing with state management
- **Context Sharing**: Data flow between steps and branches
- **Progress Tracking**: Step completion monitoring and reporting
- **Error Handling**: Comprehensive error capture and categorization

### Advanced Conditional Logic
- **Dynamic Branching**: Runtime branch selection based on context data
- **Resolver Functions**: Complex decision logic for path determination  
- **Multi-path Workflows**: Parallel execution paths with different strategies
- **Branch Convergence**: Merging execution paths after conditional processing

### Sophisticated Retry Mechanisms
- **Configurable Retries**: Maximum attempts, sleep intervals, retry conditions
- **Error Pattern Matching**: Regex-based retry triggers
- **Selective Retries**: Retry specific errors vs all errors
- **Exponential Backoff**: Progressive retry delays for stability

### Comprehensive Rollback Support
- **Automatic Rollbacks**: Trigger rollbacks on step failures
- **Granular Recovery**: Step-specific rollback functions
- **State Restoration**: Return to previous known good states
- **Resource Cleanup**: Automated cleanup of failed operations

### Production-Ready Logging
- **Structured Logging**: JSON-formatted logs with consistent fields
- **Step-level Tracking**: Individual step execution logging
- **Progress Reporting**: Real-time workflow progress updates
- **Error Details**: Comprehensive error information and stack traces

## Running the Examples

Each example is self-contained and can be run independently:

```bash
# Data Processing Pipeline
cd data-pipeline && go run main.go

# E-commerce Order Processing  
cd ecommerce-order && go run main.go

# CI/CD Pipeline Orchestrator
cd cicd-pipeline && go run main.go

# Microservice Migration Tool
cd microservice-migration && go run main.go

# Cloud Resource Provisioner
cd cloud-provisioning && go run main.go
```

## Architecture Patterns

### 1. **Sequential Processing Pattern**
Best for: Data pipelines, batch processing, linear workflows
```
Step A → Step B → Step C → Step D
```

### 2. **Conditional Branching Pattern**  
Best for: Multi-path workflows, strategy selection, environment-specific logic
```
        ┌─ Branch A ─┐
Step → ─┤           ├─ → Next Step
        └─ Branch B ─┘
```

### 3. **Hierarchical Processing Pattern**
Best for: Complex workflows, nested decision trees, multi-level processing
```
Step → Branch → Sub-Branch → Sub-Sub-Branch → Convergence
```

### 4. **Parallel Execution Pattern**
Best for: Independent operations, performance optimization, resource utilization
```
Step A ─┐
Step B ─┤─ → Convergence Step
Step C ─┘
```

## Error Handling Strategies

### 1. **Fail Fast Pattern**
Immediate termination on critical errors with comprehensive rollback.

### 2. **Graceful Degradation Pattern**  
Continue processing with reduced functionality when non-critical steps fail.

### 3. **Retry with Backoff Pattern**
Progressive retry attempts with increasing delays for transient failures.

### 4. **Circuit Breaker Pattern**
Prevent cascading failures by temporarily stopping failing operations.

## Best Practices Demonstrated

### 1. **Configuration Management**
- Externalized configuration for different environments
- Validation of configuration before execution
- Runtime configuration updates and overrides

### 2. **Resource Management**
- Proper resource allocation and cleanup
- Connection pooling and lifecycle management
- Memory and performance optimization

### 3. **Observability**
- Comprehensive logging and monitoring
- Progress tracking and status reporting  
- Performance metrics and alerting

### 4. **Testing Strategies**
- Unit testing of individual steps
- Integration testing of workflows
- Chaos engineering for failure scenarios

## Business Applications

These examples demonstrate GoSteps applicability across industries:

- **Financial Services**: Transaction processing, risk assessment, compliance workflows
- **Healthcare**: Patient data processing, clinical workflows, regulatory compliance
- **Manufacturing**: Supply chain management, quality control, production scheduling
- **Retail**: Order fulfillment, inventory management, customer service workflows
- **Technology**: Software deployment, infrastructure management, data processing

## Getting Started

1. **Choose a relevant example** based on your use case
2. **Study the implementation** to understand GoSteps patterns
3. **Adapt the patterns** to your specific requirements
4. **Extend with custom steps** and business logic
5. **Add monitoring and alerting** for production deployment

## Contributing

To add new examples or improve existing ones:

1. Follow the established patterns and documentation standards
2. Include comprehensive README with business context
3. Demonstrate multiple GoSteps features in realistic scenarios
4. Add proper error handling and rollback mechanisms
5. Include logging and monitoring capabilities

These examples showcase the versatility and power of GoSteps for building robust, maintainable, and scalable workflow orchestration systems.