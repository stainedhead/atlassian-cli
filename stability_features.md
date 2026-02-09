# Stability Features PRD

## Document Metadata

| Field          | Value                                              |
|----------------|----------------------------------------------------|
| Author         | Engineering                                        |
| Status         | Draft                                              |
| Created        | 2026-02-09                                         |
| Last Updated   | 2026-02-09                                         |
| Target Version | v0.2.0 - v0.4.0                                    |

---

## 1. Problem Statement

An audit of the current atlassian-cli codebase reveals four categories of instability that prevent production use:

1. **No token validation** -- `auth login` stores credentials without verifying them against the Atlassian REST API. Users discover invalid tokens only when a subsequent command fails, producing confusing error messages.
2. **Volatile credential storage** -- The sole `TokenManager` implementation (`MemoryTokenManager`) holds credentials in process memory. Credentials evaporate when the process exits, forcing re-authentication on every invocation.
3. **Fully mocked Confluence client** -- Every method in `internal/confluence/client.go` returns hardcoded data. The `page` and `space` commands are non-functional against real Atlassian instances.
4. **Partially mocked JIRA client** -- `ListProjects` and `GetProject` in `internal/jira/client.go` return hardcoded data instead of calling the Atlassian API. Additionally, `ListProjects` contains a variable-shadowing bug where the local `cache` variable shadows the `cache` package import.

### Secondary Issues Discovered During Audit

| Issue | Location | Severity |
|-------|----------|----------|
| `retry.contains()` only checks string prefix, not substring | `internal/retry/retry.go:76-78` | Medium |
| `isRetryable()` uses naive string matching instead of typed errors | `internal/retry/retry.go:68-74` | Medium |
| Cache has no file-level locking; concurrent access can corrupt entries | `internal/cache/cache.go` | High (blocks concurrency) |
| Audit logger writes to file without mutex; concurrent writes interleave | `internal/audit/audit.go:48-62` | High (blocks concurrency) |
| `SaveConfig` writes token in plaintext to YAML on disk | `internal/config/config.go:85-90` | High (security) |
| Viper global singleton used across all commands; not goroutine-safe | `cmd/root.go`, all command files | High (blocks concurrency) |
| YAML output format returns `not yet implemented` error | `cmd/issue/issue.go:341,375` | Low |

---

## 2. Goals

- Every stored credential is validated before being persisted.
- Credentials survive process restarts via secure, OS-native storage.
- All JIRA and Confluence commands execute real API calls.
- The architecture supports safe concurrent operation for future worker-thread / subagent patterns.

## 3. Non-Goals

- OAuth 2.0 / SAML / SSO flows (API token auth only for now).
- Interactive browser-based login.
- Multi-tenancy (running commands against multiple instances in a single invocation).
- A GUI or TUI.

---

## 4. Phased Delivery Plan

### Phase 1: Token Validation and Auth Hardening

**Objective:** Ensure credentials are verified against the Atlassian API before storage and surface clear, actionable errors on failure.

#### 4.1.1 Token Validation on Login

**Current behavior:** `cmd/auth/auth.go` calls `tokenManager.Store()` without any API call. Validation is limited to format checks (valid URL, valid email, non-empty token).

**Required behavior:**

1. After format validation passes, issue a lightweight API call to verify the token:
   - JIRA: `GET /rest/api/3/myself` (returns the authenticated user's profile).
   - Confluence: `GET /wiki/rest/api/user/current` (returns the current user).
2. If the API call fails with `401 Unauthorized`, return a specific error: `"Authentication failed: invalid email or API token. Generate a new token at https://id.atlassian.com/manage/api-tokens"`.
3. If the API call fails with a network error, return: `"Cannot reach <serverURL>. Check the URL and your network connection."`.
4. If the API call succeeds, display the authenticated user's `displayName` from the response (not just the email that was typed in).
5. Only call `tokenManager.Store()` after successful validation.

**New `auth validate` subcommand:**

Add `atlassian-cli auth validate --server <url>` that re-validates stored credentials without requiring the user to re-enter them. This enables health-check scripts and CI pipelines.

**Changes required:**

| File | Change |
|------|--------|
| `internal/auth/auth.go` | Add `Validate(ctx, serverURL, email, token) (*UserInfo, error)` to `TokenManager` interface. Add `ValidateCredentials` function that calls `/rest/api/3/myself`. |
| `cmd/auth/auth.go` | Call `Validate` before `Store` in `newLoginCmd`. Add `newValidateCmd`. |
| `internal/types/config.go` | Add `UserInfo` struct with `AccountID`, `DisplayName`, `Email`, `Active` fields. |

**Testing strategy:**
- Unit tests with `httptest.NewServer` returning 200, 401, and 5xx responses.
- Integration test struct that satisfies `TokenManager` and records call sequences.

#### 4.1.2 Fix Retry Logic

**Current bug:** `internal/retry/retry.go:76-78` -- `contains()` only checks if the error string starts with the substring:

```go
func contains(s, substr string) bool {
    return len(s) >= len(substr) && s[:len(substr)] == substr
}
```

This means an error like `"failed to connect: connection refused"` would NOT match `"connection"` because the error doesn't start with that word.

**Fix:** Replace with `strings.Contains(s, substr)`.

**Additionally:** Replace string-based error classification in `isRetryable()` with typed error checking:
- Check for `net.Error` (timeout, temporary).
- Check HTTP status codes: retry on 429, 502, 503, 504.
- Never retry on 400, 401, 403, 404, 409.

**Changes required:**

| File | Change |
|------|--------|
| `internal/retry/retry.go` | Replace `contains` with `strings.Contains`. Refactor `isRetryable` to accept `error` and inspect typed errors / HTTP status codes. Add `RetryableError` and `NonRetryableError` wrapper types. |

---

### Phase 2: Persistent Credential Storage

**Objective:** Credentials persist across CLI invocations using secure, OS-native storage with a file-based encrypted fallback.

#### 4.2.1 Keychain Token Manager

Implement `KeychainTokenManager` satisfying the existing `TokenManager` interface.

**Platform support:**

| Platform | Backend | Go Library |
|----------|---------|------------|
| macOS | Keychain Services | `github.com/keybase/go-keychain` or `github.com/zalando/go-keyring` |
| Linux | Secret Service API (GNOME Keyring / KWallet) | `github.com/zalando/go-keyring` |
| Windows | Windows Credential Manager | `github.com/zalando/go-keyring` |

**Recommendation:** Use `github.com/zalando/go-keyring` -- it provides a unified API across all three platforms with a single import.

**Storage scheme:**
- **Service name:** `atlassian-cli`
- **Account key:** The `serverURL` (e.g., `https://mycompany.atlassian.net`)
- **Secret value:** JSON-encoded `AuthCredentials` struct (email + token)

**Fallback strategy:**
If the OS keychain is unavailable (headless server, CI environment, Docker container), fall back to an encrypted file at `~/.atlassian-cli/credentials.enc`. Encrypt using AES-256-GCM with a key derived from a machine-specific identifier (hostname + user UID) via PBKDF2. This is less secure than a hardware-backed keychain but avoids plaintext storage.

**Selection logic in `cmd/root.go`:**

```
1. Try KeychainTokenManager
2. If keychain init fails, try EncryptedFileTokenManager
3. If both fail, use MemoryTokenManager and warn the user
```

**Changes required:**

| File | Change |
|------|--------|
| `internal/auth/keychain.go` | New file. Implement `KeychainTokenManager`. |
| `internal/auth/encrypted_file.go` | New file. Implement `EncryptedFileTokenManager`. |
| `cmd/root.go` | Replace `NewMemoryTokenManager()` with tiered selection logic. |
| `internal/config/config.go` | Remove `Token` field from `SaveConfig` output (stop writing plaintext tokens to YAML). |
| `go.mod` | Add `github.com/zalando/go-keyring`. |

**Testing strategy:**
- `KeychainTokenManager`: integration tests gated behind a `// +build integration` tag (require real keychain access).
- `EncryptedFileTokenManager`: unit tests using a temp directory. Verify encryption by reading the raw file and confirming it is not valid JSON.
- Mock-based unit tests for the tiered selection logic.

#### 4.2.2 Remove Plaintext Token from Config File

**Current behavior:** `internal/config/config.go:85-90` writes the API token directly into `~/.atlassian-cli/config.yaml`.

**Required behavior:**
- `SaveConfig` must never write the `Token` field to the YAML file.
- `LoadConfig` should continue to read `Token` from YAML for backwards compatibility, but emit a deprecation warning: `"Warning: token found in config file. Run 'atlassian-cli auth login' to migrate to secure storage."`.
- After successful migration, the config file entry should be removed.

---

### Phase 3: Real API Implementations

**Objective:** Replace all mock/hardcoded implementations with real Atlassian REST API calls.

#### 4.3.1 Confluence Client -- Real Implementation

**Current state:** `internal/confluence/client.go` is a `MockConfluenceClient`. All five methods return static data.

**Required changes:**

Replace `MockConfluenceClient` with a real implementation using the `go-atlassian` Confluence v2 library (`github.com/ctreminiom/go-atlassian/confluence`). The `go-atlassian` module already exists in `go.mod`.

| Method | API Endpoint | Notes |
|--------|-------------|-------|
| `CreatePage` | `POST /wiki/api/v2/pages` | Map `CreatePageRequest` to `models.PageCreatePayloadScheme`. Handle `parentID` via ancestor array. |
| `GetPage` | `GET /wiki/api/v2/pages/{id}` | Request body expansion for `body.storage`, `version`, `space`. |
| `UpdatePage` | `PUT /wiki/api/v2/pages/{id}` | Must increment version number. Fetch current version first, then update. |
| `ListPages` | `GET /wiki/api/v2/spaces/{spaceKey}/pages` or CQL search | Support pagination via `cursor` (v2 API uses cursor-based, not offset-based). |
| `ListSpaces` | `GET /wiki/api/v2/spaces` | Map to existing `SpaceListResponse` type. |

**Type mapping considerations:**
- The Confluence v2 API uses cursor-based pagination, but `PageListResponse` currently uses `StartAt`/`MaxResults` (offset-based). Either:
  - (A) Add a `Cursor` field to `PageListResponse` and update the `page list` command to support `--cursor`, OR
  - (B) Translate cursor pagination to offset semantics internally (less efficient but backwards-compatible).
  - **Recommendation:** Option A. The CLI is pre-release; breaking the pagination model now avoids tech debt.

**Changes required:**

| File | Change |
|------|--------|
| `internal/confluence/client.go` | Complete rewrite. Replace `MockConfluenceClient` with `AtlassianConfluenceClient` backed by `go-atlassian` Confluence v2. |
| `internal/types/page.go` | Add `Cursor` to `PageListResponse` and `PageListOptions`. |
| `cmd/page/page.go` | Update `newListCmd` to support `--cursor` flag. Adjust output to show cursor-based pagination info. |
| `cmd/space/space.go` | Update pagination similarly. |

#### 4.3.2 JIRA Client -- Complete Real Implementations

**Current state:** `ListProjects` and `GetProject` in `internal/jira/client.go` return hardcoded data. `ListProjects` also contains a variable-shadowing bug at line 348/382 where `cache` (the local variable) shadows the `cache` package.

**Required changes:**

| Method | API Endpoint | Notes |
|--------|-------------|-------|
| `ListProjects` | `GET /rest/api/3/project/search` | Use `c.client.Project.Search()`. Remove hardcoded mock data. Fix cache variable shadowing. |
| `GetProject` | `GET /rest/api/3/project/{projectKeyOrId}` | Use `c.client.Project.Get()`. Map response to `types.Project`. |

**Additional JIRA gaps to address:**
- `UpdateIssue` returns an error for status changes (`"status updates not yet implemented - use transitions"`). Implement transition support via `POST /rest/api/3/issue/{issueKey}/transitions`.

**Changes required:**

| File | Change |
|------|--------|
| `internal/jira/client.go` | Rewrite `ListProjects` and `GetProject` with real API calls. Fix cache shadowing bug. Add `TransitionIssue` method. |
| `internal/jira/transitions.go` | New file. Implement `GetTransitions` and `DoTransition`. |
| `internal/types/issue.go` | Add `Transition` type. |

---

### Phase 4: Concurrency Safety and Subagent Readiness

**Objective:** Make the architecture safe for concurrent command execution, whether from multiple goroutines within a single process (subagent/worker-thread model) or from multiple CLI processes accessing shared state.

#### 4.4.1 Concurrency Blockers -- Analysis

The following shared-state issues **must** be resolved before any worker-thread or subagent pattern can be safely adopted:

| Blocker | Severity | Root Cause | Impact |
|---------|----------|------------|--------|
| **Viper global singleton** | Critical | All commands use `viper.GetString()` and `viper.SetDefault()` which mutate/read global state. Two goroutines running commands concurrently will race on Viper's internal map. | Data races, corrupted config reads, unpredictable behavior. |
| **File-based cache without locking** | Critical | `cache.Set()` and `cache.Get()` perform uncoordinated read/write to `~/.atlassian-cli/cache/*.json`. | Corrupted cache files, partial reads, stale data returned as valid. |
| **Audit logger without synchronization** | High | `audit.Logger.Log()` calls `file.WriteString()` without a mutex. Concurrent goroutines writing to the same `*os.File` will interleave bytes. | Garbled audit log entries, potential data loss. |
| **Config file concurrent writes** | High | `config.SaveConfig()` uses `viper.WriteConfig()` without file locking. Multiple processes saving config simultaneously corrupt the file. | Config file corruption, lost settings. |
| **Client-per-command construction** | Medium | Each command in `cmd/issue/issue.go`, `cmd/page/page.go`, etc. constructs a new JIRA/Confluence client. No connection pooling or HTTP transport reuse. | Excessive TCP connections under concurrent load, slower performance, potential socket exhaustion. |

#### 4.4.2 Viper Isolation

**Problem:** Viper's default global instance is not goroutine-safe. The current codebase uses both `viper.GetString()` (global) and local `viper.New()` instances inconsistently.

**Solution:** Eliminate all use of the Viper global. Pass an explicit `*viper.Viper` instance through the command tree:

1. Create a single `*viper.Viper` instance in `cmd/root.go`.
2. Attach it to the Cobra command via `cmd.SetContext()` with a context key.
3. Each command retrieves its config from the context rather than calling `viper.GetString()`.

This makes each command tree self-contained and allows multiple command trees to run concurrently with independent configuration.

**Changes required:**

| File | Change |
|------|--------|
| `cmd/root.go` | Create local `*viper.Viper`, store in context. |
| `cmd/issue/issue.go` | Replace all `viper.GetString()` calls with context-based config access. |
| `cmd/page/page.go` | Same. |
| `cmd/project/project.go` | Same. |
| `cmd/space/space.go` | Same. |
| `cmd/config/config.go` | Same. |
| `internal/config/resolver.go` | Accept `*viper.Viper` parameter instead of using global. |

#### 4.4.3 Thread-Safe Cache

**Problem:** `internal/cache/cache.go` reads and writes files without any locking. Two goroutines calling `Set("projects_list", ...)` concurrently will corrupt the JSON file.

**Solution -- multi-layer approach:**

1. **In-process locking:** Add a `sync.RWMutex` per cache key (use a `sync.Map` to avoid a global lock).
2. **Cross-process locking:** Use `flock` (advisory file locking) on the cache file. On macOS/Linux, use `syscall.Flock`. On Windows, use `LockFileEx`.
3. **Atomic writes:** Write to a temp file, then `os.Rename` to the target path. This prevents partial reads.

**Changes required:**

| File | Change |
|------|--------|
| `internal/cache/cache.go` | Add `sync.RWMutex` map, atomic write via temp file + rename, `flock` for cross-process safety. |

#### 4.4.4 Thread-Safe Audit Logger

**Problem:** `audit.Logger.Log()` writes to `*os.File` without synchronization.

**Solution:**
1. Add a `sync.Mutex` to `Logger`.
2. Lock before `WriteString`, unlock after.
3. Consider buffered writes with periodic flush to reduce lock contention under high concurrency.

**Changes required:**

| File | Change |
|------|--------|
| `internal/audit/audit.go` | Add `sync.Mutex` to `Logger` struct. Lock/unlock in `Log()`. |

#### 4.4.5 Shared HTTP Client Pool

**Problem:** Each command constructs a new `v3.Client` / Confluence client, each with its own `http.Client` and TCP connection pool. Under concurrent workloads (e.g., a subagent running `issue list` and `page list` in parallel), this wastes sockets and skips HTTP/2 multiplexing.

**Solution:** Introduce a `ClientFactory` that caches and reuses clients per `(serverURL, email)` tuple:

```go
type ClientFactory struct {
    mu      sync.RWMutex
    jira    map[string]*jira.AtlassianJiraClient
    confl   map[string]*confluence.AtlassianConfluenceClient
    transport *http.Transport // shared, connection-pooling transport
}
```

All commands obtain clients from the factory rather than constructing them directly.

**Changes required:**

| File | Change |
|------|--------|
| `internal/client/factory.go` | New file. Implement `ClientFactory`. |
| `cmd/root.go` | Create `ClientFactory`, attach to context. |
| `cmd/issue/issue.go` | Obtain JIRA client from factory. |
| `cmd/page/page.go` | Obtain Confluence client from factory. |
| `cmd/project/project.go` | Obtain JIRA client from factory. |
| `cmd/space/space.go` | Obtain Confluence client from factory. |

---

## 5. Concurrency Blockers -- Summary for Subagent/Worker-Thread Architecture

If the goal is to run multiple CLI operations as concurrent worker threads (e.g., a parent orchestrator dispatching `issue list`, `page list`, and `project list` in parallel), the following blockers exist **in priority order**:

### Critical (must fix before any concurrency)

1. **Viper global state** (Phase 4.4.2) -- Without this fix, any two concurrent commands will race on shared Viper maps. This is the single largest blocker. Fix first.
2. **File-based cache without locking** (Phase 4.4.3) -- Concurrent cache writes will corrupt JSON files on disk.

### High (must fix before production concurrency)

3. **Audit logger data races** (Phase 4.4.4) -- Concurrent log writes produce garbled output.
4. **Config file corruption** (Phase 4.4.2, same fix) -- Concurrent config saves corrupt YAML.

### Medium (performance and resource management)

5. **Client-per-command construction** (Phase 4.4.5) -- Wastes TCP connections, no HTTP/2 multiplexing. Functional but inefficient.

### Safe as-is

- `MemoryTokenManager` -- already uses `sync.RWMutex`. Thread-safe.
- `KeychainTokenManager` (once built) -- `go-keyring` delegates to OS APIs that are inherently process-safe. Thread-safe.
- Individual `go-atlassian` client methods -- the underlying `http.Client` is goroutine-safe. Concurrent calls to `client.Issue.Get()` etc. are fine once the client is constructed.

### Recommended Concurrency-Safe Invocation Pattern

Once all Phase 4 work is complete, a subagent runner could safely do:

```go
factory := client.NewClientFactory(transport)
ctx := context.WithValue(ctx, configKey, viperInstance)

var wg sync.WaitGroup
wg.Add(3)

go func() {
    defer wg.Done()
    jiraClient, _ := factory.GetJiraClient(serverURL, email, token)
    issues, _ := jiraClient.ListIssues(ctx, opts)
    // process issues
}()

go func() {
    defer wg.Done()
    confClient, _ := factory.GetConfluenceClient(serverURL, email, token)
    pages, _ := confClient.ListPages(ctx, opts)
    // process pages
}()

go func() {
    defer wg.Done()
    jiraClient, _ := factory.GetJiraClient(serverURL, email, token)
    projects, _ := jiraClient.ListProjects(ctx, opts)
    // process projects
}()

wg.Wait()
```

---

## 6. Phase Dependencies

```
Phase 1: Token Validation          (no dependencies, start immediately)
Phase 2: Persistent Storage        (depends on Phase 1 -- validation must work before persisting)
Phase 3: Real API Implementations  (depends on Phase 1 -- needs working auth to test against real APIs)
Phase 4: Concurrency Safety        (depends on Phase 3 -- all clients must be real before optimizing sharing)
```

**Parallel work opportunities:**
- Phase 1 and Phase 3 can run concurrently across separate branches if Phase 1 merges first (Phase 3 feature branches can rebase onto Phase 1).
- Within Phase 3, the JIRA fixes (4.3.2) and Confluence rewrite (4.3.1) can be developed by separate engineers / subagents in parallel -- they touch disjoint files.
- Within Phase 4, cache locking (4.4.3) and audit locking (4.4.4) can be developed in parallel -- they touch disjoint files. Viper isolation (4.4.2) must come first since it affects all command files.

**Cannot be parallelized:**
- Phase 4.4.2 (Viper isolation) touches every command file. No other Phase 4 work that modifies command files can run concurrently with it.
- Phase 4.4.5 (client factory) depends on Phase 3 completing the real client implementations, and on Phase 4.4.2 completing the context-passing pattern.

---

## 7. Acceptance Criteria

### Phase 1
- [ ] `auth login` with an invalid token returns a clear error message without storing credentials.
- [ ] `auth login` with a valid token displays the user's `displayName` and stores credentials.
- [ ] `auth validate` re-checks stored credentials and reports status.
- [ ] `retry.isRetryable` correctly classifies HTTP 429, 502, 503, 504 as retryable and 400, 401, 403, 404 as non-retryable.
- [ ] All existing tests pass. New tests cover validation success, 401 failure, network failure, and retry classification.

### Phase 2
- [ ] Credentials persist after process exit and are available on the next invocation.
- [ ] On macOS, credentials are stored in Keychain (verified via `security find-generic-password`).
- [ ] On Linux, credentials are stored in Secret Service (verified via `secret-tool lookup`).
- [ ] Fallback to encrypted file works in headless environments.
- [ ] Plaintext token is no longer written to `config.yaml`.
- [ ] Existing `config.yaml` files with tokens trigger a migration warning.

### Phase 3
- [ ] `page create`, `page get`, `page list`, `page update` execute real Confluence API calls.
- [ ] `space list` executes a real Confluence API call.
- [ ] `project list` and `project get` execute real JIRA API calls.
- [ ] `issue update --status` performs a transition instead of returning "not implemented".
- [ ] The `ListProjects` cache-shadowing bug is fixed.
- [ ] No mock/hardcoded data remains in any client implementation.

### Phase 4
- [ ] Running two commands concurrently (via goroutines in a test harness) produces no data races (`go test -race` passes).
- [ ] Cache files are not corrupted under concurrent access.
- [ ] Audit log entries are well-formed under concurrent writes.
- [ ] No command uses the Viper global; `grep -r "viper\." cmd/` shows only local instance usage.
- [ ] HTTP connections are reused across concurrent commands to the same server.

---

## 8. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| `go-keyring` behaves inconsistently across Linux distros | Medium | Users unable to store credentials | Encrypted file fallback (Phase 2). Test on Ubuntu, Fedora, Arch in CI. |
| `go-atlassian` Confluence v2 API coverage is incomplete | Medium | Some Confluence operations may require raw HTTP calls | Audit `go-atlassian` v2 coverage before starting Phase 3. Fall back to direct HTTP via the library's raw request support. |
| Viper isolation is a large refactor touching every command file | High | Merge conflicts, regressions | Do this first in Phase 4 as a single PR. Run full test suite. No other Phase 4 PRs should be in-flight. |
| Keychain access may be denied in containerized CI environments | High | CI cannot run auth tests | Gate keychain tests behind `// +build integration`. CI uses `EncryptedFileTokenManager` or `MemoryTokenManager`. |
| Atlassian API rate limits during integration testing | Medium | Flaky tests, CI failures | Use HTTP record/replay (`go-vcr`) for integration tests. Reserve live API tests for a nightly suite. |

---

## 9. Estimated Scope

| Phase | Files Modified | Files Created | Approximate Complexity |
|-------|---------------|---------------|----------------------|
| Phase 1 | 4 | 0 | Small -- validation call + retry fix |
| Phase 2 | 3 | 2 | Medium -- keychain integration + encrypted fallback |
| Phase 3 | 5 | 1 | Large -- full Confluence rewrite + JIRA completion |
| Phase 4 | 12+ | 1 | Large -- refactor across entire command tree |
