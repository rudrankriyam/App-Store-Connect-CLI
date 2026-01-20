# ASC CLI - Project Plan

## Vision

A fast, AI-agent-friendly CLI for App Store Connect that enables developers to ship iOS apps with zero friction.

**Problem:** Manual App Store Connect workflows are slow, and there’s no AI‑agent‑friendly CLI with clean JSON output.

**Solution:** A lightweight Go-based CLI focused on speed, simplicity, and AI-first design.

---

## Current Reality (v0.1 - Implemented, Validated Locally)

**Last Updated:** 2026-01-20

### What Doesn't Work Yet

- **Pagination** - Manual pagination only (`--limit`, `--next`); no auto-paging yet
- **Tests** - Integration tests are opt-in and require real credentials

---

## Roadmap

### Phase 3: App Management (v0.2)

**Goal:** Add commands for managing apps and builds

#### Technical Tasks

- [ ] Add build expiration workflow

---

### Phase 3.5: Build Uploads + Submission (API 4.1+)

**Goal:** Upload builds and submit them for review without Xcode/Transporter.

#### References (Required Reading)

Use the AI-readable Apple docs via sosumi.ai (replace `developer.apple.com` with `sosumi.ai`):

- App Store Connect API 4.1 release notes (Build Uploads API)
  - https://sosumi.ai/documentation/appstoreconnectapi/app-store-connect-api-4-1-release-notes
- App Store Connect API 4.2 release notes
  - https://sosumi.ai/documentation/appstoreconnectapi/app-store-connect-api-4-2-release-notes
- App Store Connect API release notes index
  - https://sosumi.ai/documentation/appstoreconnectapi/app-store-connect-api-release-notes
- Official OpenAPI spec download (Apple Developer → Releases → App Store Connect API → Download file)
  - https://developer.apple.com/news/releases/
- Direct OpenAPI spec zip (Apple generic link used by the spec mirror)
  - https://developer.apple.com/sample-code/app-store-connect/app-store-connect-openapi-specification.zip
- OpenAPI mirror for diffing (unofficial, helpful for browsing versions)
  - https://github.com/EvanBacon/App-Store-Connect-OpenAPI-Spec

#### Target CLI (explicit, non-interactive)

1. **Upload a build**
   ```bash
   asc builds upload --app "APP_ID" --ipa "/path/to/app.ipa"
   asc builds upload --app "APP_ID" --ipa "/path/to/app.ipa" --wait
   ```

2. **Submit a build for review**
   ```bash
   asc submit --app "APP_ID" --build "BUILD_ID" --confirm
   asc submit status --id "SUBMISSION_ID"
   asc submit cancel --id "SUBMISSION_ID" --confirm
   ```

#### API Discovery Results (OpenAPI Spec 4.1)

**Build Upload Endpoints:**

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/buildUploads` | Create build upload record |
| GET | `/v1/buildUploads/{id}` | Get upload status |
| DELETE | `/v1/buildUploads/{id}` | Cancel/delete upload |
| POST | `/v1/buildUploadFiles` | Reserve upload slot, get presigned URLs |
| GET | `/v1/buildUploadFiles/{id}` | Get file upload info |
| PATCH | `/v1/buildUploadFiles/{id}` | Commit upload (set uploaded=true + checksum) |

**Submission Endpoints (App Store version submissions):**

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/appStoreVersionSubmissions` | Create submission for review |
| GET | `/v1/appStoreVersionSubmissions/{id}` | Get submission status |
| DELETE | `/v1/appStoreVersionSubmissions/{id}` | Cancel submission |

**Note:** `reviewSubmissions` is for **Game Center content** (introduced in API 4.2) and does **not** replace `appStoreVersionSubmissions` for app version review.

**BuildUploadCreateRequest Schema:**
```json
{
  "data": {
    "type": "buildUploads",
    "attributes": {
      "cfBundleShortVersionString": "1.0.0",  // required
      "cfBundleVersion": "123",                 // required
      "platform": "IOS"                         // required: IOS|MAC_OS|TV_OS|VISION_OS
    },
    "relationships": {
      "app": { "data": { "type": "apps", "id": "APP_ID" } }
    }
  }
}
```

**BuildUploadFileCreateRequest Schema:**
```json
{
  "data": {
    "type": "buildUploadFiles",
    "attributes": {
      "fileName": "app.ipa",          // required
      "fileSize": 1024000,            // required
      "uti": "com.apple.ipa",         // required: com.apple.ipa for .ipa
      "assetType": "ASSET"            // required: ASSET|ASSET_DESCRIPTION|ASSET_SPI
    },
    "relationships": {
      "buildUpload": { "data": { "type": "buildUploads", "id": "UPLOAD_ID" } }
    }
  }
}
```

**BuildUploadFileUpdateRequest Schema (Commit):**
```json
{
  "data": {
    "type": "buildUploadFiles",
    "id": "FILE_ID",
    "attributes": {
      "sourceFileChecksums": {
        "file": { "hash": "SHA256_HASH", "algorithm": "SHA_256" }
      },
      "uploaded": true
    }
  }
}
```

**AppStoreVersionSubmissionCreateRequest Schema:**
```json
{
  "data": {
    "type": "appStoreVersionSubmissions",
    "relationships": {
      "appStoreVersion": { "data": { "type": "appStoreVersions", "id": "APP_STORE_VERSION_ID" } }
    }
  }
}
```

**BuildUploadState Values (confirm in spec):**
- Common values include `AWAITING_UPLOAD`, `PROCESSING`, `FAILED`, `COMPLETE`

**DeliveryFileUploadOperation (Presigned URL):**
```json
{
  "method": "PUT",
  "url": "https://...presigned-url...",
  "length": 1000000,
  "offset": 0,
  "requestHeaders": [{"name": "Content-Length", "value": "1000000"}],
  "expiration": "2026-01-20T12:00:00Z"
}
```

**Checksum Algorithm:** `MD5 | SHA_256`

#### Upload Flow (Verified Sequence)

1. **POST /v1/buildUploads** - Create upload record, get `UPLOAD_ID`
2. **POST /v1/buildUploadFiles** - Reserve slot, get presigned URLs in `uploadOperations`
3. **PUT** to each `uploadOperation.url` - Upload file parts in order
4. **PATCH /v1/buildUploadFiles/{id}** - Set `uploaded: true` + checksums to commit
5. **GET /v1/buildUploads/{id}** - Poll `state` until `COMPLETE` or `FAILED`
6. Extract `build.id` from response relationships for submission

#### Submission Flow (Verified Sequence)

1. **Find or create App Store Version** for the app/platform
   - `GET /v1/apps/{id}/appStoreVersions?filter[platform]=IOS` (and optionally filter by state)
2. **Attach build to App Store Version**
   - `PATCH /v1/appStoreVersions/{id}/relationships/build` with `builds/{id}`
3. **POST /v1/appStoreVersionSubmissions** - Create submission (returns `SUBMISSION_ID`)
4. **GET /v1/appStoreVersionSubmissions/{id}** - Check `state` until submitted/in review/complete
5. **DELETE /v1/appStoreVersionSubmissions/{id}`** - Cancel (guarded by `--confirm`)

#### Safety & Guardrails

- Add `--confirm` for submission and cancellation
- Add `--wait` with bounded timeout (default off)
- Provide clear errors for “metadata not ready” / “agreements missing”
- Never attempt submission in tests unless explicitly opted-in

#### TDD Checklist (Implementation)

**Client layer**
- [ ] HTTP tests for each new endpoint (method, path, body schema)
- [ ] Upload operations tests: verify headers and payload order
- [ ] Error parsing for failed upload/submit responses

**CLI layer**
- [ ] Validation tests (missing `--app`, `--ipa`, `--confirm`)
- [ ] Output tests (json/table/markdown)
- [ ] JSON remains minified by default

**Integration tests (opt‑in)**
- [ ] `ASC_UPLOAD_APP_ID` (required)
- [ ] `ASC_UPLOAD_IPA_PATH` (required)
- [ ] `ASC_CONFIRM_SUBMIT=true` (required for submit)
- [ ] `ASC_SUBMIT_BUILD_ID` (optional if upload returns build ID)
- [ ] Skip if any env vars are missing

#### Data to Capture with Live Credentials

Non‑destructive calls to record example response shapes:
- `asc apps`
- `asc builds --app "APP_ID"`
- `asc builds info --build "BUILD_ID"`
- `asc reviews --app "APP_ID"`

Use these to verify:
- Field names (e.g., `uploadedDate`, `processingState`, `expired`)
- Status strings returned by the API

---

### Phase 4: Beta Management (v0.3)

**Goal:** Add commands for managing beta testers and groups

#### Features

1. **Beta Groups**
   ```bash
   asc beta-groups list --app "APP_ID"
   asc beta-groups create --app "APP_ID" --name "Beta Testers"
   ```

2. **Beta Testers**
   ```bash
   asc beta-testers list --app "APP_ID"
   asc beta-testers add --app "APP_ID" --email "tester@example.com" --group "Beta"
   asc beta-testers remove --app "APP_ID" --email "tester@example.com"
   ```

3. **Tester Invitations**
   ```bash
   asc beta-testers invite --app "APP_ID" --email "tester@example.com"
   ```

#### Technical Tasks

- [ ] Implement `GET /v1/apps/{id}/betaGroups`
- [ ] Implement `POST /v1/betaGroups`
- [ ] Implement `GET /v1/apps/{id}/betaTesters`
- [ ] Implement `POST /v1/betaTesters`
- [ ] Implement `DELETE /v1/betaTesters/{id}`

---

### Phase 5: Localization (v0.4)

**Goal:** Add commands for managing app localizations

#### Features

1. **Download Localizations**
   ```bash
   asc localizations download --app "APP_ID" --output ./locales/
   ```

2. **Upload Localizations**
   ```bash
   asc localizations upload --app "APP_ID" --path ./locales/
   ```

3. **List Localizations**
   ```bash
   asc localizations list --app "APP_ID"
   ```

#### Technical Tasks

- [ ] Implement `GET /v1/apps/{id}/appStoreVersions`
- [ ] Implement `GET /v1/apps/{id}/appStoreVersionLocalizations`
- [ ] Implement file upload/download for localization files
- [ ] Add validation for localization files

---

### Phase 6: Submission (v0.5)

**Goal:** Add commands for submitting apps

#### Features

1. **Submit for Review**
   ```bash
   asc submit --app "APP_ID" --build "BUILD_ID" --submit-type "APP_STORE"
   ```

2. **Check Submission Status**
   ```bash
   asc submit status --app "APP_ID"
   ```

3. **Cancel Submission**
   ```bash
   asc submit cancel --app "APP_ID"
   ```

#### Technical Tasks

- [ ] Implement `POST /v1/appStoreVersionSubmissions`
- [ ] Implement `GET /v1/appStoreVersionSubmissions/{id}`
- [ ] Implement `DELETE /v1/appStoreVersionSubmissions/{id}`
- [ ] Add workflow for submission process

---

### Phase 7: Sandbox & Testing (v0.6)

**Goal:** Add commands for sandbox testing

#### Features

1. **Create Sandbox Tester**
   ```bash
   asc sandbox create --email "tester@example.com" --territory US
   ```

2. **List Sandbox Testers**
   ```bash
   asc sandbox list
   ```

3. **Delete Sandbox Tester**
   ```bash
   asc sandbox delete --email "tester@example.com"
   ```

#### Technical Tasks

- [ ] Implement `GET /v1/sandboxTesters`
- [ ] Implement `POST /v1/sandboxTesters`
- [ ] Implement `DELETE /v1/sandboxTesters/{id}`

---

### Phase 8: Analytics (v0.7)

**Goal:** Add commands for viewing analytics

#### Features

1. **Sales Report**
   ```bash
   asc analytics sales --app "APP_ID" --from 2024-01-01 --to 2024-01-31
   ```

2. **Usage Report**
   ```bash
   asc analytics usage --app "APP_ID" --last-30-days
   ```

#### Technical Tasks

- [ ] Implement `GET /v1/salesReports`
- [ ] Implement `GET /v1/usageMetrics`
- [ ] Add report formatting

---

### Future Enhancements (v1.0+)

- **Interactive Mode** - TUI for manual exploration
- **Plugins** - Extendable architecture
- **AI Summarization** - Use local LLMs to summarize feedback
- **Auto-Responder** - AI-powered response to reviews
- **Multi-Account** - Manage multiple ASC accounts
- **Web Dashboard** - Optional web UI

---

## Dependencies

### Core Dependencies

```
github.com/peterbourgon/ff/v3     - CLI framework
github.com/golang-jwt/jwt/v5      - JWT signing
```

### Dev Tools (optional)

```
golangci-lint                     - Linting (if installed)
gosec                             - Security scanning (if installed)
github.com/goreleaser/nfpm/v2     - Packaging via `go run` (optional)
```

---

## Release Strategy

1. **Alpha** - Internal testing, core features
2. **Beta** - Public testing, feedback collection
3. **v0.1** - First public release (feedback, crashes, reviews)
4. **v0.x** - Incremental feature releases
5. **v1.0** - Full feature set, stable API

---

## Success Metrics

- Install via Homebrew: `brew install rudrank/tap/asc`
- Average startup time: < 50ms
- JSON output is default; use `--output` for table/markdown
- 80%+ test coverage
- Zero security vulnerabilities

---

## Current Status

Next: Implement build uploads + submission flows (API 4.1+), then auto-pagination and beta management

## Known Issues

1. **Pagination**
   - Manual pagination only (`--limit`, `--next`); no auto-paging yet

2. **Filtering**
   - Reviews support rating/territory filters
   - Feedback/crashes support device model, OS version, platform, build, and tester filters

3. **Keychain**
   - Keychain supported; local config fallback still exists

