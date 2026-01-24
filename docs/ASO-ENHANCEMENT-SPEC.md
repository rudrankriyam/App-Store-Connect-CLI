# ASC CLI Fork Specification
## App Store Connect CLI Enhancement Plan

**Goal:** Extend the [App-Store-Connect-CLI](https://github.com/rudrankriyam/App-Store-Connect-CLI) to support full ASO (App Store Optimization) and metadata management, enabling complete replacement of fastlane for App Store content management.

**Approach:** Fork → Implement → PR upstream

---

## Table of Contents
1. [Current CLI Capabilities](#current-cli-capabilities)
2. [Priority 1: Core ASO Features](#priority-1-core-aso-features)
3. [Priority 2: Asset Management](#priority-2-asset-management)
4. [Priority 3: Version & Release Management](#priority-3-version--release-management)
5. [Priority 4: In-App Purchases & Subscriptions](#priority-4-in-app-purchases--subscriptions)
6. [Priority 5: Custom Product Pages & Experiments](#priority-5-custom-product-pages--experiments)
7. [Priority 6: Advanced Features](#priority-6-advanced-features)
8. [Fastlane Migration Feature](#fastlane-migration-feature)
9. [Claude Code Skill Integration](#claude-code-skill-integration)
10. [Implementation Notes](#implementation-notes)

---

## Current CLI Capabilities

### Already in Main Branch ✅

| Category | Features | Commands |
|----------|----------|----------|
| **Localizations** | Full metadata CRUD (description, keywords, whatsNew, promotionalText, supportUrl, marketingUrl) | `localizations list/download/upload` |
| **App Info** | Name, subtitle, privacy URLs | `localizations --type app-info` |
| **Versions** | List, get, attach build | `versions list/get/attach-build` |
| **Phased Release** | Create, pause, resume, complete | `versions phased-release get/create/update/delete` |
| **Submit** | Submit for review, check status, cancel | `submit create/status/cancel` |
| **TestFlight** | Beta groups, testers, feedback, crash reports | `beta-groups`, `beta-testers`, `feedback`, `crashes` |
| **Reviews** | Customer reviews, review responses | `reviews list`, `reviews responses` |
| **Analytics** | Sales summaries, analytics reports, financial reports | `analytics`, `finance` |
| **Sandbox** | Sandbox testers, purchase history | `sandbox` |
| **Xcode Cloud** | Workflows, build runs, polling | `xcode-cloud` |
| **Apps/Builds** | List apps, builds | `apps list`, `builds list/latest` |

### In Upstream Feature Branches (Not Merged) 🔄

| Branch | Features | Lines Added |
|--------|----------|-------------|
| `cursor/app-store-assets-management-e5a2` | Screenshot/preview upload, asset management | +2,889 |
| `cursor/app-store-review-submissions-c76d` | Enhanced review submissions | ~500 |
| `cursor/in-app-purchases-and-subscriptions-5fad` | IAP and subscription management | TBD |
| `cursor/app-pricing-and-availability-bc84` | Pricing and territory availability | TBD |

### Actually Needs Implementation 🔨

| Feature | Priority | Notes |
|---------|----------|-------|
| **Version CREATE** | P1 | Only list/get/attach-build exists |
| **Categories** | P1 | Primary/secondary category management |
| **Fastlane Migration** | P1 | Import/export fastlane directory structure |
| **Custom Product Pages** | P2 | Marketing campaign pages |
| **In-App Events** | P2 | Event cards and scheduling |
| **A/B Experiments** | P3 | Product page optimization |
| **Age Rating** | P3 | Content rating declarations |
| **App Privacy** | P3 | Privacy nutrition labels |

### Strategy: Merge vs Build

For features already in upstream branches, we should:
1. **Review** the existing branch implementation
2. **Merge** if it meets our needs, or **enhance** if gaps exist
3. **Focus new work** on: version create, categories, fastlane migration

---

## Priority 1: Core ASO Features

**Rationale:** These are the fundamental features needed to manage App Store listings and replace fastlane's `deliver` action.

### 1.1 App Store Version Localizations

Manage localized metadata per version (description, keywords, What's New, etc.)

```bash
# List localizations for a version
asc metadata list --app-id <id> --version <version>
asc metadata list --app-id <id> --version latest

# Get specific locale
asc metadata get --app-id <id> --version <version> --locale en-US

# Update localized fields
asc metadata update --app-id <id> --version <version> --locale en-US \
  --description "New app description" \
  --keywords "keyword1,keyword2,keyword3" \
  --whats-new "Bug fixes and improvements" \
  --promotional-text "Now with dark mode!"

# Update from file
asc metadata update --app-id <id> --version <version> --locale en-US \
  --description-file ./metadata/en-US/description.txt

# Bulk update from directory (fastlane-compatible structure)
asc metadata sync --app-id <id> --version <version> --metadata-dir ./metadata
```

**API Endpoints:**
- `GET /v1/appStoreVersions/{id}/appStoreVersionLocalizations`
- `GET /v1/appStoreVersionLocalizations/{id}`
- `PATCH /v1/appStoreVersionLocalizations/{id}`
- `POST /v1/appStoreVersionLocalizations`

**Fields:**
| Field | Max Length | Notes |
|-------|------------|-------|
| `description` | 4,000 chars | Main app description |
| `keywords` | 100 chars | Comma-separated |
| `whatsNew` | 4,000 chars | Release notes |
| `promotionalText` | 170 chars | Can update without new version |
| `marketingUrl` | URL | Optional |
| `supportUrl` | URL | Required |
| `privacyPolicyUrl` | URL | May be required |

### 1.2 App Info Localizations

Manage app-level metadata (name, subtitle) that persists across versions.

```bash
# List app info localizations
asc app-info list --app-id <id>

# Update app name and subtitle
asc app-info update --app-id <id> --locale en-US \
  --name "My App Name" \
  --subtitle "The best app ever"

# Privacy policy text (if applicable)
asc app-info update --app-id <id> --locale en-US \
  --privacy-choices-url "https://..."
```

**API Endpoints:**
- `GET /v1/apps/{id}/appInfos`
- `GET /v1/appInfos/{id}/appInfoLocalizations`
- `PATCH /v1/appInfoLocalizations/{id}`

**Fields:**
| Field | Max Length | Notes |
|-------|------------|-------|
| `name` | 30 chars | App name on store |
| `subtitle` | 30 chars | Below app name |
| `privacyPolicyText` | - | For Kids category |
| `privacyChoicesUrl` | URL | Privacy nutrition label |
| `privacyPolicyUrl` | URL | Required for some apps |

### 1.3 Categories

Manage primary and secondary app categories.

```bash
# List available categories
asc categories list

# Get current categories for app
asc app-info categories --app-id <id>

# Update categories
asc app-info update-categories --app-id <id> \
  --primary "HEALTH_AND_FITNESS" \
  --primary-subcategory "HEALTH_AND_FITNESS_YOGA" \
  --secondary "LIFESTYLE"
```

**API Endpoints:**
- `GET /v1/appCategories`
- `PATCH /v1/appInfos/{id}` (with category relationships)

---

## Priority 2: Asset Management

**Rationale:** Screenshots and previews are essential for ASO but involve complex multi-step upload workflows.

### 2.1 Screenshot Management

```bash
# List screenshot sets for a version/locale
asc screenshots list --app-id <id> --version <version> --locale en-US

# List screenshots in a set
asc screenshots list --app-id <id> --version <version> --locale en-US \
  --display-type APP_IPHONE_67

# Upload screenshot
asc screenshots upload --app-id <id> --version <version> --locale en-US \
  --display-type APP_IPHONE_67 \
  --file ./screenshots/iphone67_01.png

# Upload all screenshots from directory
asc screenshots sync --app-id <id> --version <version> \
  --screenshots-dir ./screenshots

# Delete screenshot
asc screenshots delete --screenshot-id <id>

# Reorder screenshots
asc screenshots reorder --app-id <id> --version <version> --locale en-US \
  --display-type APP_IPHONE_67 \
  --order "id1,id2,id3"
```

**Display Types:**
| Type | Device |
|------|--------|
| `APP_IPHONE_67` | iPhone 15 Pro Max (6.7") |
| `APP_IPHONE_61` | iPhone 15 Pro (6.1") |
| `APP_IPHONE_65` | iPhone 11 Pro Max (6.5") |
| `APP_IPHONE_58` | iPhone X/XS (5.8") |
| `APP_IPHONE_55` | iPhone 8 Plus (5.5") |
| `APP_IPAD_PRO_129` | iPad Pro 12.9" |
| `APP_IPAD_PRO_3GEN_129` | iPad Pro 12.9" (3rd gen) |
| `APP_IPAD_105` | iPad Pro 10.5" |
| `APP_APPLE_WATCH_SERIES_7` | Apple Watch Series 7+ |
| `APP_APPLE_TV` | Apple TV |
| `APP_DESKTOP` | macOS |

**Upload Process (3-step):**
1. **Reserve:** `POST /v1/appScreenshots` → returns upload operations
2. **Upload:** PUT chunks to presigned URLs (parallel)
3. **Commit:** `PATCH /v1/appScreenshots/{id}` with MD5 checksum

### 2.2 App Preview Management

```bash
# List preview sets
asc previews list --app-id <id> --version <version> --locale en-US

# Upload preview video
asc previews upload --app-id <id> --version <version> --locale en-US \
  --display-type APP_IPHONE_67 \
  --file ./previews/demo.mp4 \
  --poster-frame-time 5.0

# Delete preview
asc previews delete --preview-id <id>
```

**Preview Requirements:**
- Format: H.264, AAC audio
- Duration: 15-30 seconds
- Resolution: Must match device display
- File size: Up to 500MB

### 2.3 App Icon

```bash
# Get current app icon
asc app-icon get --app-id <id> --version <version>

# Note: App icon is typically uploaded with the build, not separately
```

---

## Priority 3: Version & Release Management

**Rationale:** Essential for the complete submission workflow.

### 3.1 Version Management

```bash
# List versions
asc versions list --app-id <id>
asc versions list --app-id <id> --platform IOS --state READY_FOR_SALE

# Create new version
asc versions create --app-id <id> --platform IOS --version-string "2.1.0"

# Get version details
asc versions get --version-id <id>

# Update version
asc versions update --version-id <id> \
  --copyright "2026 My Company" \
  --release-type MANUAL

# Delete version (if not submitted)
asc versions delete --version-id <id>
```

**Version States:**
| State | Description |
|-------|-------------|
| `PREPARE_FOR_SUBMISSION` | Editing allowed |
| `WAITING_FOR_REVIEW` | In queue |
| `IN_REVIEW` | Being reviewed |
| `PENDING_DEVELOPER_RELEASE` | Approved, awaiting manual release |
| `READY_FOR_SALE` | Live on App Store |
| `DEVELOPER_REJECTED` | Rejected by developer |
| `REJECTED` | Rejected by Apple |
| `METADATA_REJECTED` | Metadata issues |

### 3.2 Build Association

```bash
# List builds eligible for version
asc builds list --app-id <id> --version <version> --eligible

# Associate build with version
asc versions set-build --version-id <id> --build-id <build-id>
```

### 3.3 App Review Submission

```bash
# Submit for review
asc submit --app-id <id> --version <version>

# Submit with options
asc submit --app-id <id> --version <version> \
  --release-type AFTER_APPROVAL \
  --uses-idfa false

# Cancel submission
asc submit cancel --app-id <id> --version <version>

# Check submission status
asc submit status --app-id <id> --version <version>
```

**Submission Information:**
```bash
# Set review contact info
asc review-info update --app-id <id> --version <version> \
  --contact-first-name "John" \
  --contact-last-name "Doe" \
  --contact-email "john@example.com" \
  --contact-phone "+1-555-0100" \
  --demo-account-name "demo@example.com" \
  --demo-account-password "password123" \
  --notes "Use the demo account to test premium features"

# Add review attachment
asc review-info add-attachment --app-id <id> --version <version> \
  --file ./review-guide.pdf
```

### 3.4 Phased Release

```bash
# Enable phased release
asc phased-release create --version-id <id>

# Get phased release status
asc phased-release status --version-id <id>

# Pause phased release
asc phased-release pause --version-id <id>

# Resume phased release
asc phased-release resume --version-id <id>

# Complete phased release (release to all)
asc phased-release complete --version-id <id>
```

**Phased Release Schedule:**
| Day | Percentage |
|-----|------------|
| 1 | 1% |
| 2 | 2% |
| 3 | 5% |
| 4 | 10% |
| 5 | 20% |
| 6 | 50% |
| 7 | 100% |

### 3.5 Release Management

```bash
# Manual release (for PENDING_DEVELOPER_RELEASE)
asc release --app-id <id> --version <version>

# Schedule release
asc release --app-id <id> --version <version> \
  --scheduled-date "2026-02-01T10:00:00Z"
```

---

## Priority 4: In-App Purchases & Subscriptions

**Rationale:** Important for monetization management, though some configuration still requires the web UI.

### 4.1 In-App Purchase Management

```bash
# List in-app purchases
asc iap list --app-id <id>
asc iap list --app-id <id> --type CONSUMABLE

# Get IAP details
asc iap get --iap-id <id>

# Create in-app purchase
asc iap create --app-id <id> \
  --type NON_CONSUMABLE \
  --product-id "com.example.premium" \
  --reference-name "Premium Upgrade"

# Update IAP
asc iap update --iap-id <id> \
  --cleared-for-sale true

# IAP localizations
asc iap localization update --iap-id <id> --locale en-US \
  --display-name "Premium Upgrade" \
  --description "Unlock all features"
```

**IAP Types:**
- `CONSUMABLE`
- `NON_CONSUMABLE`
- `NON_RENEWING_SUBSCRIPTION`

### 4.2 Subscription Management

```bash
# List subscription groups
asc subscriptions groups list --app-id <id>

# List subscriptions in group
asc subscriptions list --group-id <id>

# Get subscription details
asc subscriptions get --subscription-id <id>

# Update subscription
asc subscriptions update --subscription-id <id> \
  --group-level 1

# Subscription localizations
asc subscriptions localization update --subscription-id <id> --locale en-US \
  --display-name "Monthly Pro" \
  --description "Full access for one month"
```

### 4.3 Pricing

```bash
# List price points
asc pricing list-points --territory US

# Get subscription price equalizations
asc pricing equalizations --price-point-id <id>

# Set subscription price
asc subscriptions set-price --subscription-id <id> \
  --base-territory US \
  --price-point <price-point-id>

# Set app price
asc app pricing set --app-id <id> \
  --price-point <price-point-id>

# List territories
asc territories list
```

### 4.4 Availability

```bash
# Get app availability
asc availability get --app-id <id>

# Update app availability
asc availability update --app-id <id> \
  --available-in-new-territories true

# Set specific territories
asc availability set-territories --app-id <id> \
  --territories US,CA,GB,AU

# IAP availability
asc iap availability set --iap-id <id> \
  --territories US,CA,GB
```

---

## Priority 5: Custom Product Pages & Experiments

**Rationale:** Advanced ASO features for marketing optimization.

### 5.1 Custom Product Pages

```bash
# List custom product pages
asc cpp list --app-id <id>

# Create custom product page
asc cpp create --app-id <id> \
  --name "Holiday Campaign" \
  --url-suffix "holiday2026"

# Get CPP details
asc cpp get --cpp-id <id>

# CPP versions
asc cpp versions list --cpp-id <id>
asc cpp versions create --cpp-id <id> --app-version <version-id>

# CPP localizations (screenshots, promotional text)
asc cpp localization update --cpp-version-id <id> --locale en-US \
  --promotional-text "Special holiday offer!"

# Submit CPP for review
asc cpp submit --cpp-version-id <id>
```

### 5.2 Product Page Optimization (A/B Testing)

```bash
# List experiments
asc experiments list --app-id <id>

# Create experiment
asc experiments create --app-id <id> \
  --name "Icon Test Q1" \
  --traffic-proportion 0.5

# Experiment treatments
asc experiments treatments list --experiment-id <id>

# Update treatment (different icon/screenshots)
asc experiments treatments update --treatment-id <id> \
  --app-icon ./icons/variant_b.png

# Start experiment
asc experiments start --experiment-id <id>

# Stop experiment
asc experiments stop --experiment-id <id>

# Get experiment results
asc experiments results --experiment-id <id>
```

### 5.3 In-App Events

```bash
# List in-app events
asc events list --app-id <id>

# Create in-app event
asc events create --app-id <id> \
  --reference-name "Summer Sale" \
  --badge "LIVE_EVENT" \
  --event-state READY_FOR_REVIEW

# Event localizations
asc events localization update --event-id <id> --locale en-US \
  --name "Summer Sale Event" \
  --short-description "50% off all items" \
  --long-description "Join our biggest sale of the year..."

# Event media
asc events media upload --event-id <id> --locale en-US \
  --type EVENT_CARD \
  --file ./events/summer-card.png

# Submit event for review
asc events submit --event-id <id>
```

---

## Priority 6: Advanced Features

**Rationale:** Nice-to-have features for complete coverage.

### 6.1 Pre-Orders

```bash
# Enable pre-order
asc pre-order enable --app-id <id> \
  --release-date "2026-03-15"

# Get pre-order status
asc pre-order status --app-id <id>

# Update pre-order release date
asc pre-order update --app-id <id> \
  --release-date "2026-03-20"

# Cancel pre-order
asc pre-order cancel --app-id <id>
```

### 6.2 Age Rating

```bash
# Get age rating declaration
asc age-rating get --app-id <id>

# Update age rating
asc age-rating update --app-id <id> \
  --violence-cartoon NONE \
  --violence-realistic INFREQUENT_OR_MILD \
  --gambling false \
  --alcohol-tobacco-drugs NONE
```

### 6.3 App Privacy

```bash
# List privacy declarations
asc privacy list --app-id <id>

# Update privacy declarations
asc privacy update --app-id <id> \
  --data-types "NAME,EMAIL_ADDRESS,PHONE_NUMBER" \
  --purposes "APP_FUNCTIONALITY,ANALYTICS"
```

### 6.4 Game Center

```bash
# List leaderboards
asc game-center leaderboards list --app-id <id>

# List achievements
asc game-center achievements list --app-id <id>

# Note: Full Game Center configuration may require web UI
```

### 6.5 App Clips

```bash
# List app clip experiences
asc app-clips list --app-id <id>

# Update app clip
asc app-clips update --clip-id <id> \
  --action OPEN
```

---

## Fastlane Migration Feature

**Goal:** Seamlessly import existing fastlane metadata structure and provide compatibility commands.

### 7.1 Import Command

```bash
# Import fastlane metadata directory
asc migrate import --app-id <id> --version <version> \
  --fastlane-dir ./fastlane

# Dry run (preview changes)
asc migrate import --app-id <id> --version <version> \
  --fastlane-dir ./fastlane \
  --dry-run

# Import only metadata (no screenshots)
asc migrate import --app-id <id> --version <version> \
  --fastlane-dir ./fastlane \
  --metadata-only

# Import only screenshots
asc migrate import --app-id <id> --version <version> \
  --fastlane-dir ./fastlane \
  --screenshots-only
```

### 7.2 Export Command

```bash
# Export current App Store metadata to fastlane format
asc migrate export --app-id <id> --version <version> \
  --output-dir ./fastlane

# Export for comparison/backup
asc migrate export --app-id <id> --version latest \
  --output-dir ./backup
```

### 7.3 Fastlane Directory Structure Support

The CLI should recognize and work with fastlane's directory structure:

```
fastlane/
├── metadata/
│   ├── copyright.txt
│   ├── primary_category.txt
│   ├── secondary_category.txt
│   ├── primary_first_sub_category.txt
│   ├── primary_second_sub_category.txt
│   ├── secondary_first_sub_category.txt
│   ├── secondary_second_sub_category.txt
│   ├── review_information/
│   │   ├── demo_password.txt
│   │   ├── demo_user.txt
│   │   ├── email_address.txt
│   │   ├── first_name.txt
│   │   ├── last_name.txt
│   │   ├── notes.txt
│   │   └── phone_number.txt
│   ├── default/                    # Default locale values
│   │   ├── keywords.txt
│   │   ├── marketing_url.txt
│   │   ├── privacy_url.txt
│   │   └── support_url.txt
│   └── en-US/                      # Per-locale
│       ├── description.txt
│       ├── keywords.txt
│       ├── marketing_url.txt
│       ├── name.txt
│       ├── privacy_url.txt
│       ├── promotional_text.txt
│       ├── release_notes.txt
│       ├── subtitle.txt
│       └── support_url.txt
└── screenshots/
    ├── en-US/
    │   ├── iPhone 15 Pro Max-01.png
    │   ├── iPhone 15 Pro Max-02.png
    │   ├── iPad Pro (12.9-inch)-01.png
    │   └── ...
    └── de-DE/
        └── ...
```

### 7.4 Filename-to-Display-Type Mapping

Support fastlane's screenshot naming conventions:

| Filename Pattern | Display Type |
|------------------|--------------|
| `*iPhone 15 Pro Max*` or `*IPHONE_67*` | `APP_IPHONE_67` |
| `*iPhone 15 Pro*` or `*IPHONE_61*` | `APP_IPHONE_61` |
| `*iPhone 8 Plus*` or `*IPHONE_55*` | `APP_IPHONE_55` |
| `*iPad Pro (12.9-inch)*` or `*IPAD_PRO_129*` | `APP_IPAD_PRO_129` |
| `*iPad Pro (12.9-inch) (3rd*` | `APP_IPAD_PRO_3GEN_129` |
| `*Apple Watch*` | `APP_APPLE_WATCH_SERIES_7` |
| `*Apple TV*` | `APP_APPLE_TV` |

---

## Claude Code Skill Integration

### 8.1 SKILL.md Content

```markdown
# App Store Connect CLI Skill

You have access to the `asc` CLI for managing App Store Connect.

## Common Operations

### View App Metadata
```bash
asc metadata list --app-id $APP_ID --version latest
```

### Update Description
```bash
asc metadata update --app-id $APP_ID --version latest --locale en-US \
  --description "Your new description"
```

### Update Keywords
```bash
asc metadata update --app-id $APP_ID --version latest --locale en-US \
  --keywords "keyword1,keyword2,keyword3"
```

### Upload Screenshots
```bash
asc screenshots sync --app-id $APP_ID --version latest \
  --screenshots-dir ./screenshots
```

### Create New Version
```bash
asc versions create --app-id $APP_ID --platform IOS --version-string "2.0.0"
```

### Submit for Review
```bash
asc submit --app-id $APP_ID --version "2.0.0"
```

### Check Version Status
```bash
asc versions list --app-id $APP_ID --state PENDING_DEVELOPER_RELEASE
```

### Release Approved Version
```bash
asc release --app-id $APP_ID --version "2.0.0"
```

## Environment

Set these environment variables for authentication:
- `ASC_KEY_ID`: API Key ID
- `ASC_ISSUER_ID`: Issuer ID
- `ASC_PRIVATE_KEY_PATH`: Path to .p8 private key
- `ASC_APP_ID`: Default app ID (optional)
```

### 8.2 Workflow Examples

**ASO Update Workflow:**
```
User: "Update the keywords for Stitch It to focus on alterations"

Claude: I'll update the keywords for your app.
[Executes: asc metadata update --app-id <id> --locale en-US --keywords "alterations,tailoring,clothing repair,hem,resize"]
```

**Release Workflow:**
```
User: "Create version 2.1.0 and prepare it for submission"

Claude: I'll create the new version and set it up.
[Executes: asc versions create --app-id <id> --platform IOS --version-string "2.1.0"]
[Executes: asc versions set-build --version-id <new-id> --build-id <latest-build>]
[Executes: asc metadata update --app-id <id> --version "2.1.0" --locale en-US --whats-new "Bug fixes"]
```

---

## Implementation Notes

### API Authentication

The existing CLI uses keychain storage with config file fallback. No changes needed.

### Rate Limiting

Apple's API typically allows 50 requests/minute with burst capacity. Implement:
- Exponential backoff on 429 responses
- Request queuing for bulk operations
- Progress reporting for long uploads

### Error Handling

Map API error codes to user-friendly messages:
| Code | Meaning | User Message |
|------|---------|--------------|
| 409 | Conflict | "Version is in a state that doesn't allow this operation" |
| 403 | Forbidden | "API key doesn't have permission for this operation" |
| 404 | Not Found | "Resource not found - check ID" |

### Screenshot Upload Implementation

The 3-step upload process requires:
1. MD5 checksum calculation
2. Parallel chunk uploads to presigned URLs
3. Commit with checksum verification

Consider using goroutines for parallel uploads with progress reporting.

### Testing Strategy

1. **Unit tests:** Mock API responses
2. **Integration tests:** Use sandbox/test app
3. **Manual testing:** Test with real apps before PR

---

## Implementation Phases (Revised)

### Phase 0: Branch Integration
- [ ] Review `cursor/app-store-assets-management-e5a2` branch
- [ ] Review `cursor/in-app-purchases-and-subscriptions-5fad` branch
- [ ] Review `cursor/app-pricing-and-availability-bc84` branch
- [ ] Merge or cherry-pick relevant work into our fork

### Phase 1: Core Gaps (MVP)
- [ ] `versions create` - Create new App Store versions
- [ ] `versions update` - Update copyright, release type
- [ ] `versions delete` - Delete draft versions
- [ ] Categories management (`categories list`, `app-info update-categories`)
- [ ] Review info management (demo account, contact info)

### Phase 2: Fastlane Migration
- [ ] `migrate import` - Import from fastlane metadata/ directory
- [ ] `migrate export` - Export to fastlane format
- [ ] Filename-to-display-type mapping for screenshots
- [ ] Support for `default/` locale fallback

### Phase 3: Assets (if not covered by branch)
- [ ] Verify screenshot upload works from merged branch
- [ ] Add directory sync with fastlane naming conventions
- [ ] App preview (video) upload support

### Phase 4: Advanced ASO
- [ ] Custom product pages (create, localize, submit)
- [ ] In-app events (create, localize, schedule)
- [ ] Product page A/B experiments

### Phase 5: Polish & PR
- [ ] Claude Code SKILL.md creation
- [ ] Comprehensive documentation
- [ ] Test suite for new commands
- [ ] PR upstream with clear commit history

---

## References

- [App Store Connect API Documentation](https://developer.apple.com/documentation/appstoreconnectapi)
- [App Metadata API](https://developer.apple.com/documentation/appstoreconnectapi/app-metadata)
- [App Store Version Localizations](https://developer.apple.com/documentation/appstoreconnectapi/app_store/app_metadata/app_store_version_localizations)
- [Screenshot Upload Guide (Runway)](https://www.runway.team/blog/how-to-upload-assets-using-the-app-store-connect-api)
- [Fastlane Deliver Documentation](https://docs.fastlane.tools/actions/deliver/)
- [Existing CLI Repository](https://github.com/rudrankriyam/App-Store-Connect-CLI)
