# GitHub Workflows

This directory contains GitHub Actions workflows for automated testing, building, and deployment validation of the Molecule application.

## Workflows Overview

### Core CI/CD Workflows

- **`ci.yaml`** - Comprehensive continuous integration with unit tests, linting, building, and security scanning
- **`build-and-push.yaml`** - Build and push Docker images to GitHub Container Registry (existing)
- **`golangci-lint.yaml`** - Go code linting (existing)
- **`release.yaml`** - Create GitHub releases from tags (existing)

### Integration and Testing Workflows

- **`integration.yaml`** - Integration tests with mock Nomad server and load testing
- **`docker-integration.yaml`** - Docker build testing, multi-architecture builds, and security scanning
- **`e2e.yaml`** - End-to-end tests with real Nomad cluster and Docker Compose scenarios
- **`performance.yaml`** - Performance benchmarks, memory usage, and startup time testing

### Deployment Workflows

- **`deployment-validation.yaml`** - Deployment readiness checks, smoke tests, and configuration validation

## Workflow Triggers

- **On Push/PR**: `ci.yaml`, `integration.yaml`, `docker-integration.yaml`, `e2e.yaml`
- **On Tags**: `build-and-push.yaml`, `release.yaml`, `deployment-validation.yaml` 
- **Scheduled**: `integration.yaml` (daily), `performance.yaml` (weekly)
- **Manual**: `e2e.yaml`, `deployment-validation.yaml`

## Test Coverage

The workflows provide comprehensive testing coverage:

1. **Unit Tests** - All Go packages with race detection and coverage reporting
2. **Integration Tests** - Mock Nomad deployments with API testing
3. **Load Testing** - Performance under concurrent requests
4. **Security Scanning** - Code and Docker image vulnerability scanning  
5. **End-to-End Testing** - Full application deployment scenarios
6. **Performance Testing** - Memory usage, startup time, and benchmarks

## Mock Deployment Testing

The integration workflows include mock deployment scenarios that:

- Start a real Nomad server in a container
- Deploy test applications to Nomad
- Start Molecule with test configuration
- Test all API endpoints and functionality
- Validate performance and resource usage
- Test error handling and edge cases

This ensures that Molecule works correctly in production-like environments.