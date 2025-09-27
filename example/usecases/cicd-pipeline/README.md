# CI/CD Pipeline Orchestrator

This example demonstrates a comprehensive CI/CD pipeline using GoSteps that showcases advanced workflow orchestration, conditional branching based on build characteristics, and robust error handling with rollback mechanisms.

## Features Demonstrated

### 1. **Multi-Stage Pipeline Architecture**

- **Pipeline Initialization**: Environment setup and configuration
- **Source Code Management**: Git checkout with network error handling
- **Build Process**: Artifact generation and compilation
- **Testing Strategy**: Dynamic test suite selection
- **Security Scanning**: Vulnerability assessment with retry logic
- **Deployment Orchestration**: Environment-specific deployment strategies
- **Notification System**: Team communication and status updates

### 2. **Dynamic Testing Strategies**

Based on the `testSuite` configuration, the pipeline automatically selects:

**Unit Tests Only** (`testSuite: "unit"`)
- Fast feedback for development builds
- Code coverage reporting
- Basic quality gates

**Integration Testing** (`testSuite: "integration"`)
- Unit tests + Integration tests
- Database and service integration validation
- Environment dependency testing

**Full Test Suite** (`testSuite: "all"` or `"e2e"`)
- Complete testing pyramid: Unit → Integration → E2E
- End-to-end user journey validation
- Production-like environment testing

### 3. **Intelligent Deployment Strategies**

The deployment resolver automatically determines the strategy based on:

**Staging Deployment** (Non-production environments)
- Quick deployment for development/testing
- Basic health checks
- Simplified rollback procedures

**Production Deployment** (Standard production releases)
- Robust deployment with comprehensive verification
- Production smoke tests
- Enhanced monitoring and alerting

**Canary Deployment** (Production releases with `buildType: "release"`)
- Gradual traffic shifting (1% → 100%)
- Real-time metrics monitoring
- Automatic rollback on performance degradation
- Risk mitigation for critical releases

### 4. **Comprehensive Error Handling**

**Network-Related Errors**
- Git checkout timeouts and network failures
- Retry with exponential backoff
- Automatic workspace cleanup on failures

**Build and Test Failures**
- Compilation errors with artifact cleanup
- Flaky test handling with intelligent retries
- Test environment provisioning issues

**Deployment Failures**
- Automatic rollback to previous stable version
- Infrastructure provisioning errors
- Service health check failures

### 5. **Advanced Rollback Mechanisms**

- **Workspace Cleanup**: Removes temporary files and build artifacts
- **Test Environment Reset**: Cleans up test databases and services
- **Docker Image Cleanup**: Removes failed container builds
- **Deployment Rollback**: Reverts to previous stable deployment
- **Artifact Cleanup**: Removes incomplete or corrupted build outputs

## Pipeline Architecture

```
┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│ Initialize Pipeline │ -> │   Checkout Code     │ -> │  Build Application  │
└─────────────────────┘    └─────────────────────┘    └─────────────────────┘
                                                                │
                                                      ┌─────────▼─────────┐
                                                      │    Run Tests      │
                                                      │   (Resolver)      │
                                                      └─────────┬─────────┘
                                       ┌──────────────────────┼──────────────────────┐
                                       │                      │                      │
                              ┌────────▼────────┐    ┌────────▼────────┐    ┌────────▼────────┐
                              │ Unit Tests Only │    │ Integration     │    │ Full Test Suite │
                              │                 │    │ Tests           │    │                 │
                              │ • Unit Tests    │    │ • Unit Tests    │    │ • Unit Tests    │
                              └─────────┬───────┘    │ • Integration   │    │ • Integration   │
                                        │            └─────────┬───────┘    │ • E2E Tests     │
                                        │                      │            └─────────┬───────┘
                                        └──────────────────────┼──────────────────────┘
                                                               │
                                                     ┌─────────▼─────────┐
                                                     │ Build Docker      │
                                                     │ Image             │
                                                     └─────────┬─────────┘
                                                               │
                                                     ┌─────────▼─────────┐
                                                     │ Security          │
                                                     │ Scanning          │
                                                     └─────────┬─────────┘
                                                               │
                                                     ┌─────────▼─────────┐
                                                     │ Deploy            │
                                                     │ Application       │
                                                     │ (Resolver)        │
                                                     └─────────┬─────────┘
                               ┌──────────────────────────────┼──────────────────────────────┐
                               │                              │                              │
                      ┌────────▼────────┐          ┌────────▼────────┐          ┌────────▼────────┐
                      │ Staging Deploy  │          │ Production      │          │ Canary Deploy   │
                      │                 │          │ Deploy          │          │                 │
                      │ • Deploy        │          │ • Deploy        │          │ • Deploy 1%     │
                      │ • Verify        │          │ • Verify        │          │ • Monitor       │
                      └─────────┬───────┘          │ • Smoke Tests   │          │ • Promote 100%  │
                                │                  └─────────┬───────┘          └─────────┬───────┘
                                └──────────────────────────┼──────────────────────────────┘
                                                           │
                                                 ┌─────────▼─────────┐
                                                 │ Notify Team       │
                                                 └───────────────────┘
```

## Configuration Examples

### Feature Branch Build
```go
Build{
    ID: "BUILD-2025-001",
    ProjectName: "microservice-api",
    Branch: "feature/user-auth",
    Environment: "staging",
    BuildType: "feature",
    TestSuite: "integration",
}
```

### Production Release Build
```go
Build{
    ID: "BUILD-2025-002", 
    ProjectName: "microservice-api",
    Branch: "main",
    Environment: "production",
    BuildType: "release",
    TestSuite: "all",
}
```

## Running the Example

```bash
cd example/usecases/cicd-pipeline
go run main.go
```

## Key Decision Points

### Test Suite Selection
- **Unit**: Fast feedback for development
- **Integration**: Service integration validation
- **All/E2E**: Complete quality assurance

### Deployment Strategy Selection
1. **Environment Check**: Production vs Non-production
2. **Build Type**: Release builds use canary deployment
3. **Risk Assessment**: Critical changes get additional safeguards

### Retry Logic Patterns
- **Git Operations**: Network timeout handling
- **Test Execution**: Flaky test mitigation
- **Security Scanning**: Service availability issues
- **Deployment**: Infrastructure provisioning delays

## Business Value

This CI/CD pipeline demonstrates:

1. **Quality Assurance**: Multi-layered testing ensures code quality
2. **Risk Mitigation**: Canary deployments reduce production risk
3. **Developer Productivity**: Fast feedback loops and automated processes
4. **Operational Excellence**: Comprehensive monitoring and rollback capabilities
5. **Scalability**: Modular design supports growing team needs

The pipeline ensures reliable software delivery while maintaining high development velocity and system stability.