# Health Package

The Health package provides system health monitoring capabilities for the GreenTrade API, allowing for both basic and detailed health checks.

## Core Components

### Health Check Endpoints

The package exposes two main endpoints:

1. **Basic Health Check**: A simple endpoint that confirms the API is running
2. **Detailed Health Check**: An advanced endpoint providing comprehensive system metrics

### System Metrics Collection

The detailed health check collects various metrics:

1. **Runtime Information**:

   - Go version
   - Number of CPU cores
   - Active goroutines

2. **Memory Statistics**:

   - Currently allocated memory
   - Total allocated memory since start
   - Memory obtained from the operating system
   - Garbage collection cycles

3. **Uptime Tracking**:

   - Application start time
   - Current uptime duration

4. **Database Connectivity**:
   - Database connection status
   - Connection latency

### Data Models

The package defines several structures for representing system information:

1. **SystemInfo**: Overall system metrics
2. **MemStats**: Memory usage statistics
3. **Health Response**: API response format for health data

## Implementation Details

The Health package implements health monitoring with these features:

1. **Zero-Dependency Metrics**: Core metrics that don't depend on external services
2. **Dependency Checks**: Verification of critical dependencies like the database
3. **Structured Logging**: Recording health check requests for monitoring
4. **Consistent Response Format**: Standardized response structure for tooling integration

## Usage in Monitoring

The health endpoints are designed to work with:

1. **Load Balancers**: To determine if an instance should receive traffic
2. **Monitoring Systems**: For ongoing health surveillance
3. **Alerting Tools**: To trigger notifications when health issues arise
