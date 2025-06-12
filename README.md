# Method

A CLI HTTP request client designed for developers and testers. Store your HTTP requests in YAML files with support for environment variables, authentication flows, value resolvers, and more. Think of it as a file-based alternative to Postman that integrates seamlessly into your development workflow.

## Installation

```bash
go install github.com/mnalsup/method/cmd/method@latest
```

Or build from source:
```bash
git clone https://github.com/mnalsup/method.git
cd method
make install
```

## User Guide

### Basic Usage

```bash
method <request-file>.yaml
```

### YAML Request File Structure

Method uses YAML files to define HTTP requests. Here's the complete structure with all available options:

```yaml
# Basic request configuration
method: POST
url: https://api.example.com/v1/users
timeoutSeconds: 30  # Optional: request timeout in seconds (default: 10)

# Headers
headers:
  Content-Type: application/json
  Accept: application/json
  User-Agent: method-cli/1.0

# Request body (choose one approach)
body:
  # Structured data (will be JSON encoded for application/json)
  name: "John Doe"
  email: "john@example.com"
  active: true

# OR raw string body
bodyStr: |
  {
    "name": "John Doe",
    "email": "john@example.com"
  }

# File uploads for multipart requests
files:
  - requestBodyPath: avatar
    filePath: /path/to/avatar.jpg
  - requestBodyPath: document
    filePath: /path/to/document.pdf

# Value resolvers for dynamic data injection
valueResolvers:
  - name: auth_token
    requestPath: auth/login.yaml
    jsonParseBodyPath: access_token
    target:
      header:
        name: Authorization
        formatString: "Bearer %s"
  
  - name: api_key
    requestPath: auth/api-key.yaml
    jsonParseBodyPath: key
    target:
      environmentVariable:
        name: API_KEY
```

### Request Body Types

Method supports different content types automatically based on the `Content-Type` header:

#### JSON Requests
```yaml
method: POST
url: https://api.example.com/users
headers:
  Content-Type: application/json
body:
  name: "Alice"
  age: 30
  tags: ["developer", "golang"]
```

#### Form-Encoded Requests
```yaml
method: POST
url: https://api.example.com/form-submit
headers:
  Content-Type: application/x-www-form-urlencoded
body:
  username: alice
  password: secret123
```

#### Multipart Form Data
```yaml
method: POST
url: https://api.example.com/upload
headers:
  Content-Type: multipart/form-data
body:
  title: "My Document"
  description: "Document upload"
files:
  - requestBodyPath: file
    filePath: /path/to/document.pdf
```

#### Raw String Body
```yaml
method: POST
url: https://api.example.com/webhook
headers:
  Content-Type: text/plain
bodyStr: "Raw text content here"
```

### Environment Variable Substitution

Use environment variables in your YAML files:

```yaml
method: GET
url: ${API_BASE_URL}/users/${USER_ID}
headers:
  Authorization: Bearer ${API_TOKEN}
  X-Environment: ${ENVIRONMENT}
```

Set environment variables before running:
```bash
export API_BASE_URL="https://api-prod.example.com"
export USER_ID="12345"
export API_TOKEN="your-token-here"
export ENVIRONMENT="production"

method get-user.yaml
```

### Value Resolvers

Value resolvers allow you to chain requests - using the response from one request to configure another. This is particularly useful for authentication flows.

#### Header Injection
```yaml
# main-request.yaml
method: GET
url: https://api.example.com/protected-data
valueResolvers:
  - name: authentication
    requestPath: auth/login.yaml
    jsonParseBodyPath: token
    target:
      header:
        name: Authorization
        formatString: "Bearer %s"
```

#### Environment Variable Injection
```yaml
valueResolvers:
  - name: session_setup
    requestPath: auth/session.yaml
    jsonParseBodyPath: session_id
    target:
      environmentVariable:
        name: SESSION_ID

body:
  sessionId: ${SESSION_ID}
  action: "get_data"
```

### Timeout Configuration

Configure request timeouts to handle slow or unresponsive services:

```yaml
method: GET
url: https://slow-api.example.com/data
timeoutSeconds: 60  # 60 second timeout

# Default timeout is 10 seconds if not specified
```

### Sample Files

Check the `sample/` directory for example request files:

- `sample/json.yaml` - Simple GET request
- `sample/structuredBody.yaml` - Complex JSON body
- `sample/timeout-example.yaml` - Timeout configuration
- `sample/example-with-value-resolver.yaml` - Value resolver usage
- `sample/envsubst.yaml` - Environment variable substitution

### Response Output

Method displays:
- HTTP status code and reason
- Response headers
- Response body (formatted JSON when applicable)
- Request duration

Example output:
```
🚀 METHOD v2.0 - Using new ValueResolver architecture
--------------------Results--------------------
200 OK
Content-Type: [application/json]
Content-Length: [156]

{
  "id": 123,
  "name": "John Doe",
  "email": "john@example.com"
}
Duration: 245.123456ms
-----------------------------------------------
```

## Developer Guide

### Building from Source

#### Prerequisites
- Go 1.17 or later
- Make (optional, for convenience commands)

#### Build Commands

```bash
# Clone the repository
git clone https://github.com/mnalsup/method.git
cd method

# Download dependencies
make deps
# or: go mod download

# Build binary
make build
# or: go build -o method ./cmd/method

# Install to $GOPATH/bin
make install
# or: go install ./cmd/method
```

#### Development Commands

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o method-linux ./cmd/method
GOOS=windows GOARCH=amd64 go build -o method.exe ./cmd/method
GOOS=darwin GOARCH=amd64 go build -o method-darwin ./cmd/method
```

#### Project Structure

```
method/
├── cmd/
│   └── method/           # Main application entry point
│       └── main.go
├── internal/             # Internal packages
│   ├── schemav1.go      # Request schema definitions
│   ├── http/            # HTTP client implementation
│   │   ├── request.go   # Request execution
│   │   ├── handler.go   # Request handling logic
│   │   └── formatter.go # Response formatting
│   └── yaml/            # YAML parsing utilities
├── logging/             # Logging utilities
├── sample/              # Example request files
├── coverage/            # Test coverage reports
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
├── Makefile            # Build automation
└── README.md           # This file
```

#### Adding New Features

1. **Request Schema Changes**: Update `internal/schemav1.go`
2. **HTTP Features**: Modify files in `internal/http/`
3. **Output Formatting**: Update `internal/http/formatter.go`
4. **New Content Types**: Add support in `internal/http/request.go`

#### Testing

```bash
# Run all tests
go test ./...

# Test specific package
go test ./internal/http

# Run tests with verbose output
go test -v ./...

# Test with race detection
go test -race ./...
```

#### Debugging

```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -o method-debug ./cmd/method

# Run with debugging
dlv exec ./method-debug -- sample/json.yaml
```

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

### Areas for Contribution

- New request/response content types
- Additional authentication strategies
- Output format options (XML, CSV, etc.)
- Performance improvements
- Better error handling and messaging

### Development Setup

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and add tests
4. Run tests: `go test ./...`
5. Commit your changes: `git commit -m 'Add amazing feature'`
6. Push to the branch: `git push origin feature/amazing-feature`
7. Open a Pull Request

## License

[Add your license information here]
