# FairOS-dfs Integration Tests

Comprehensive integration test suite covering all major features of fairOS-dfs.

## Test Files

### 1. `comprehensive_integration_test.go`
Tests core SDK functionality with realistic workflows:
- **Account Registration**: User creation, mnemonic generation, login/logout
- **Multi-User Pod Sharing**: Two-user scenarios with pod sharing and receiving
- **Sync Operations**: Session persistence, directory/file sync across open/close
- **Delete Operations**: File, directory, and pod deletion with verification
- **ACT-Based Sharing**: Secure sharing with Access Control Trie
- **File Transfer**: Multi-file pod sharing simulating file sending
- **Real-World Scenario**: Complete team collaboration workflow

### 2. `api_integration_test.go`
Tests HTTP API endpoints:
- **User API**: Signup, login, session management
- **Pod API**: Create, open, list, close, delete, share, receive
- **File API**: Upload (multipart), download, list, delete, stat
- **Directory API**: Create, list, remove
- **Full Workflow**: End-to-end user journey via API
- **Error Handling**: Invalid operations, authentication failures

## Running Tests

### Prerequisites

1. **Swarm node** (Bee) running locally:
   ```bash
   bee start --dev
   ```

2. **Go environment**:
   ```bash
   go version  # Should be 1.21+
   ```

### Run All Integration Tests

```bash
cd /Users/gregor/Data/fairOS-dfs
go test -v ./pkg/test/comprehensive_integration_test.go
go test -v ./pkg/test/api_integration_test.go
```

### Run Specific Tests

```bash
# Account registration only
go test -v ./pkg/test/comprehensive_integration_test.go -run TestComprehensiveWorkflow/1_AccountRegistration

# Sync operations only
go test -v ./pkg/test/comprehensive_integration_test.go -run TestComprehensiveWorkflow/3_SyncOperations

# API pod operations
go test -v ./pkg/test/api_integration_test.go -run TestAPIPodOperations

# Full workflow
go test -v ./pkg/test/comprehensive_integration_test.go -run TestRealWorldScenario
```

### Run with Coverage

```bash
go test -v -coverprofile=coverage.out ./pkg/test/comprehensive_integration_test.go
go tool cover -html=coverage.out -o coverage.html
```

## Test Scenarios

### TestComprehensiveWorkflow
Complete feature coverage in isolated subtests:

1. **Account Registration**
   - Creates user with mnemonic
   - Verifies address generation
   - Tests authentication

2. **Multi-User Pod Sharing**
   - User 1 creates pod and uploads files
   - User 1 shares pod
   - User 2 receives shared pod
   - Verifies access

3. **Sync Operations**
   - Creates pod with directories and files
   - Closes and reopens pod
   - Verifies all data persisted
   - Tests multiple sync cycles

4. **Delete Operations**
   - Creates test structure
   - Deletes files and directories
   - Deletes entire pod
   - Verifies cleanup

5. **ACT-Based Sharing**
   - Owner creates pod with confidential data
   - Shares with specific grantee using public key
   - Grantee receives via ACT reference
   - Verifies access control

6. **File Transfer Operations**
   - Sender uploads multiple files
   - Shares entire pod
   - Receiver accesses all files
   - Simulates data exchange

### TestRealWorldScenario
Complete team collaboration workflow:
- Project lead registers and creates project pod
- Sets up folder structure (docs/, src/, tests/)
- Uploads project files
- Team members register
- Lead shares project with team
- Sync verification across sessions
- Multi-user workflow validation

### API Tests
All major API endpoints with:
- Success cases
- Error cases
- Full request/response validation
- Session cookie management
- Multipart file uploads

## Expected Output

Successful test run shows:
```
=== RUN   TestComprehensiveWorkflow
=== Starting Comprehensive Integration Test ===

=== RUN   TestComprehensiveWorkflow/1_AccountRegistration
Testing account registration...
✓ Account created successfully
  Address: 0x1234567890
  Mnemonic length: 12 words
--- PASS: TestComprehensiveWorkflow/1_AccountRegistration (0.15s)

=== RUN   TestComprehensiveWorkflow/2_MultiUserPodSharing
Testing multi-user pod sharing scenario...
✓ Created 2 users
✓ User 1 created pod: shared-documents
✓ User 1 uploaded file: /documents/report.pdf
✓ Pod shared with reference: 9c8d7e6f5a4b3c2d1e0f...
✓ User 2 received shared pod successfully
--- PASS: TestComprehensiveWorkflow/2_MultiUserPodSharing (0.45s)

[... more test output ...]

=== All Comprehensive Integration Tests Passed ===
--- PASS: TestComprehensiveWorkflow (2.31s)
PASS
```

## Test Features

### What's Tested

✅ **Account Operations**
- User registration with mnemonic
- Login with password
- Login with mnemonic
- Session management
- Wallet export

✅ **Pod Operations**
- Create pod
- Open/close pod
- List pods
- Delete pod
- Share pod (basic and ACT-based)
- Receive shared pod

✅ **File Operations**
- Upload files (various sizes)
- Download files
- List files/directories
- Delete files
- Rename files/directories
- Stat operations

✅ **Directory Operations**
- Create directories
- List directory contents
- Delete directories (recursive)
- Rename directories

✅ **Sync Operations**
- Pod data persistence
- Cross-session sync
- Directory tree sync
- File metadata sync

✅ **Sharing & Access Control**
- Public sharing (sharing reference)
- ACT-based sharing (grantee-specific)
- Multi-user scenarios
- Access verification

✅ **Error Handling**
- Invalid credentials
- Non-existent resources
- Permission errors
- Validation failures

### What's NOT Tested

❌ Group operations (separate test suite)
❌ Subscription/billing operations
❌ Large file operations (>100MB)
❌ Concurrent access conflicts
❌ Network failure scenarios
❌ Rate limiting

## Troubleshooting

### "connection refused" errors
Ensure Bee node is running:
```bash
# Check if Bee is running
curl http://localhost:1633/health

# Start Bee in dev mode
bee start --dev
```

### "postage batch not found"
Tests use mock Bee server, shouldn't need real postage batches. If error persists, check test setup.

### Memory issues
Some tests create large files. If OOM:
```bash
# Increase Go test timeout
go test -v -timeout 30m ./pkg/test/...
```

### Flaky tests
Integration tests can be flaky due to timing. Run with:
```bash
go test -v -count=3 ./pkg/test/...  # Run 3 times
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests
on: [push, pull_request]

jobs:
  integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Start Bee
        run: |
          docker run -d -p 1633:1633 ethersphere/bee:latest start --dev
          sleep 10

      - name: Run Integration Tests
        run: |
          go test -v ./pkg/test/comprehensive_integration_test.go
          go test -v ./pkg/test/api_integration_test.go
```

## Maintenance

When adding new features:
1. Add test case to appropriate test file
2. Follow existing patterns (setup → action → verification)
3. Include cleanup in defer/teardown
4. Document new test in this README

## Performance Benchmarks

Run performance tests:
```bash
go test -bench=. -benchmem ./pkg/test/...
```

Expected performance (local dev):
- Account registration: ~100-200ms
- Pod creation: ~50-100ms
- File upload (1MB): ~100-200ms
- Sync operation: ~50-150ms

## Related Documentation

- [fairOS-dfs API Docs](https://docs.fairos.fairdatasociety.org/api/)
- [Swarm Documentation](https://docs.ethswarm.org/)
- [Go Testing Guide](https://go.dev/doc/tutorial/add-a-test)
