# Cloud Resource Provisioner

This example demonstrates a sophisticated cloud infrastructure provisioning system using GoSteps that handles complex resource dependencies, multiple provisioning strategies, and comprehensive rollback mechanisms.

## Features Demonstrated

### 1. **Intelligent Resource Dependency Management**

**Dependency Analysis**
- Automatic dependency graph construction
- Circular dependency detection and prevention
- Topological sorting for optimal provisioning order

**Resource Relationships**
- Hard dependencies (required for functionality)
- Soft dependencies (preferred but not critical)
- Cross-resource configuration dependencies

### 2. **Multi-Strategy Provisioning**

**Network-First Strategy** (Production deployments)
- Sequential provisioning prioritizing stability
- Network infrastructure → Compute → Database → Load Balancers
- Comprehensive validation at each layer
- Optimal for complex, mission-critical deployments

**Parallel Strategy** (Development/Testing)
- Concurrent provisioning for speed
- Independent resources provisioned simultaneously
- Dependent resources follow after prerequisites
- Faster deployment for simpler architectures

### 3. **Comprehensive Resource Support**

**Networking Resources**
- Virtual Private Clouds (VPCs)
- Subnets (public/private)
- Security groups and NACLs
- Route tables and gateways

**Compute Resources**
- EC2 instances with auto-scaling
- Container services (ECS/EKS)
- Lambda functions
- Elastic Beanstalk environments

**Database Resources**
- RDS instances (Multi-AZ support)
- DynamoDB tables
- ElastiCache clusters
- Redshift data warehouses

**Load Balancing**
- Application Load Balancers
- Network Load Balancers
- Target group configuration
- Health check setup

### 4. **Advanced Error Handling & Recovery**

**Provisioning Failures**
- Resource quota exceeded handling
- Service limit detection
- Network connectivity issues
- Authentication/authorization problems

**Retry Strategies**
- Exponential backoff for temporary failures
- Resource-specific retry policies
- Dependency-aware retry ordering
- Maximum attempt limitations

**Rollback Mechanisms**
- Granular resource destruction by type
- Dependency-aware cleanup ordering
- State preservation during partial failures
- Configuration reset capabilities

### 5. **Production-Ready Features**

**Authentication & Authorization**
- Multi-cloud provider support
- Credential management and rotation
- Service account authentication
- Role-based access control

**Resource Validation**
- Pre-provisioning configuration validation
- Post-provisioning health checks
- Security compliance verification
- Network connectivity testing

**Monitoring & Reporting**
- Real-time provisioning progress
- Resource status tracking
- Detailed deployment reports
- Resource inventory management

## Provisioning Architecture

```
┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│ Validate            │ -> │ Authenticate        │ -> │ Analyze             │
│ Configuration       │    │ Provider            │    │ Dependencies        │
└─────────────────────┘    └─────────────────────┘    └─────────────────────┘
                                                                │
                                                      ┌─────────▼─────────┐
                                                      │ Provision         │
                                                      │ Resources         │
                                                      │ (Strategy)        │
                                                      └─────────┬─────────┘
                                             ┌─────────────────┼─────────────────┐
                                             │                 │                 │
                                    ┌────────▼────────┐      ┌─▼───────────────┐
                                    │ Network-First   │      │ Parallel        │
                                    │ Strategy        │      │ Strategy        │
                                    │                 │      │                 │
                                    │ 1. Networking   │      │ 1. Independent  │
                                    │ 2. Compute      │      │    Resources    │
                                    │ 3. Database     │      │ 2. Dependent    │
                                    │ 4. Load Balancer│      │    Resources    │
                                    └────────┬────────┘      └─┬───────────────┘
                                             │                 │
                                             └─────────┬───────┘
                                                       │
                                              ┌────────▼────────┐
                                              │ Configure       │
                                              │ Resources       │
                                              └────────┬────────┘
                                                       │
                                              ┌────────▼────────┐
                                              │ Validate        │
                                              │ Deployment      │
                                              └────────┬────────┘
                                                       │
                                              ┌────────▼────────┐
                                              │ Generate        │
                                              │ Outputs         │
                                              └─────────────────┘
```

## Sample Infrastructure Configuration

```go
ProvisioningJob{
    ID:            "PROVISION-2025-001",
    Name:          "Production Web Application Infrastructure", 
    CloudProvider: "aws",
    Region:        "us-west-2",
    Environment:   "production",
    Resources: []CloudResource{
        // VPC for network isolation
        {
            ID:   "vpc-1",
            Name: "main-vpc",
            Type: "vpc",
            Specifications: map[string]interface{}{
                "cidr_block":           "10.0.0.0/16",
                "enable_dns_hostnames": true,
                "enable_dns_support":   true,
            },
        },
        // Public subnet for load balancers
        {
            ID:   "subnet-1", 
            Name: "public-subnet-1",
            Type: "subnet",
            Specifications: map[string]interface{}{
                "cidr_block":        "10.0.1.0/24",
                "availability_zone": "us-west-2a",
                "public":            true,
            },
            Dependencies: []string{"vpc-1"},
        },
        // Private subnet for databases
        {
            ID:   "subnet-2",
            Name: "private-subnet-1", 
            Type: "subnet",
            Specifications: map[string]interface{}{
                "cidr_block":        "10.0.2.0/24",
                "availability_zone": "us-west-2b",
                "public":            false,
            },
            Dependencies: []string{"vpc-1"},
        },
        // RDS PostgreSQL database
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
            Dependencies: []string{"subnet-2"},
        },
        // Auto-scaling web application
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
            Dependencies: []string{"subnet-1", "db-1"},
        },
        // Application load balancer
        {
            ID:   "lb-1",
            Name: "application-loadbalancer",
            Type: "loadbalancer",
            Specifications: map[string]interface{}{
                "type":   "application",
                "scheme": "internet-facing",
                "subnets": []string{"subnet-1"},
            },
            Dependencies: []string{"app-1"},
        },
    },
}
```

## Running the Example

```bash
cd example/usecases/cloud-provisioning
go run main.go
```

## Provisioning Strategies Comparison

| Strategy      | Use Case | Speed | Reliability | Resource Order |
|---------------|----------|-------|-------------|----------------|
| Network-First | Production, Complex | Slower | Higher | Sequential by type |
| Parallel      | Development, Simple | Faster | Standard | Dependency-based |

## Error Scenarios Handled

### Resource Provisioning Failures
- **Quota Limits**: Service-specific resource quotas exceeded
- **Capacity Issues**: Insufficient capacity in target regions/zones
- **Configuration Errors**: Invalid resource specifications
- **Permission Issues**: IAM/RBAC access denied

### Network & Connectivity Issues
- **Authentication Timeouts**: Cloud provider API timeouts
- **Network Partitions**: Temporary connectivity losses
- **DNS Resolution**: Service discovery failures
- **Security Groups**: Blocking required traffic

### Dependency Resolution
- **Circular Dependencies**: Automatic detection and prevention
- **Missing Dependencies**: Validation before provisioning
- **Partial Failures**: Cleanup of successfully provisioned resources
- **State Inconsistency**: Resource drift detection and correction

## Business Value

This cloud provisioning system provides:

1. **Infrastructure as Code**: Repeatable, version-controlled infrastructure
2. **Risk Mitigation**: Comprehensive validation and rollback capabilities  
3. **Cost Optimization**: Efficient resource provisioning and cleanup
4. **Operational Excellence**: Automated dependency management and monitoring
5. **Multi-Cloud Support**: Provider-agnostic resource provisioning
6. **Developer Productivity**: Simplified infrastructure deployment workflows

The system ensures reliable, efficient, and secure cloud infrastructure provisioning while maintaining operational visibility and control.