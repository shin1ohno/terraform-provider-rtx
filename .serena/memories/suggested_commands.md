# Suggested Commands for Terraform Provider RTX Development

## Build and Installation
```bash
# Build the provider
make build
# or directly:
go build -o terraform-provider-rtx

# Install provider locally for testing
make install
# This installs to ~/.terraform.d/plugins/registry.terraform.io/sh1/rtx/{version}/{os_arch}
```

## Testing Commands
```bash
# Run unit tests
make test
# or:
go test ./...

# Run unit tests with verbose output
go test -v ./...

# Run acceptance tests (requires RTX router or Docker environment)
make testacc
# or:
TF_ACC=1 go test ./... -v -timeout 120m

# Run acceptance tests with Docker test environment
docker-compose -f test/docker/docker-compose.yml up -d
export RTX_HOST=localhost
export RTX_PORT=2222
export RTX_USERNAME=admin
export RTX_PASSWORD=password
TF_ACC=1 go test ./... -v
```

## Code Quality Commands
```bash
# Format Go code
make fmt
# or:
gofmt -s -w .

# Format Terraform examples
terraform fmt -recursive ./examples/

# Run linter (golangci-lint needs to be installed)
make lint
# or:
golangci-lint run
```

## Development Commands
```bash
# Generate code (if using go:generate directives)
make generate
# or:
go generate ./...

# Generate documentation
make docs
# or:
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

# Clean build artifacts
make clean
```

## Docker Test Environment
```bash
# Start RTX simulator
cd test/docker
docker-compose up -d

# Change RTX model for testing
RTX_MODEL=RTX830 docker-compose up -d

# View logs
docker-compose logs -f

# Stop and remove
docker-compose down
```

## Module Management
```bash
# Download dependencies
go mod download

# Update dependencies
go get -u ./...

# Tidy dependencies
go mod tidy

# Vendor dependencies (if needed)
go mod vendor
```

## Debugging and Development
```bash
# Run specific test
go test -v -run TestRTXSystemInfoDataSourceRead_Success ./internal/provider/

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Check for race conditions
go test -race ./...
```

## Git Commands (Darwin/macOS specific)
```bash
# Common git operations
git status
git diff
git add .
git commit -m "message"
git push origin main
git log --oneline -10

# Create and switch branch
git checkout -b feature/new-feature

# Interactive rebase (be careful)
git rebase -i HEAD~3
```

## System Utilities (Darwin/macOS)
```bash
# File operations
ls -la
find . -name "*.go" -type f
grep -r "pattern" --include="*.go" .

# Process management
ps aux | grep terraform
lsof -i :2222  # Check what's using port 2222

# Network debugging
nc -zv localhost 2222  # Test SSH port
ssh -p 2222 admin@localhost  # Manual SSH test
```