---
name: asc-release-flow
description: End-to-end release workflows for TestFlight and App Store using asc publish, builds, versions, and submit commands. Use when asked to upload a build, distribute to TestFlight, or submit to App Store.
---

# Release flow (TestFlight and App Store)

Use this skill when you need to get a new build into TestFlight or submit to the App Store.

## Preconditions
- Ensure credentials are set (`asc auth login` or `ASC_*` env vars).
- Use a new build number for each upload.
- Prefer `ASC_APP_ID` or pass `--app` explicitly.

## Preferred end-to-end commands
- TestFlight:
  - `asc publish testflight --app <APP_ID> --ipa <PATH> --group <GROUP_ID>[,<GROUP_ID>]`
  - Optional: `--wait`, `--notify`, `--platform`, `--poll-interval`, `--timeout`
- App Store:
  - `asc publish appstore --app <APP_ID> --ipa <PATH> --version <VERSION>`
  - Optional: `--wait`, `--submit --confirm`, `--platform`, `--poll-interval`, `--timeout`

## Manual sequence (when you need more control)
1. Upload the build:
   - `asc builds upload --app <APP_ID> --ipa <PATH>`
2. Find the build ID if needed:
   - `asc builds latest --app <APP_ID> [--version <VERSION>] [--platform <PLATFORM>]`
3. TestFlight distribution:
   - `asc builds add-groups --build <BUILD_ID> --group <GROUP_ID>[,<GROUP_ID>]`
4. App Store attach + submit:
   - `asc versions attach-build --version-id <VERSION_ID> --build <BUILD_ID>`
   - `asc submit create --app <APP_ID> --version <VERSION> --build <BUILD_ID> --confirm`
5. Check or cancel submission:
   - `asc submit status --id <SUBMISSION_ID>` or `--version-id <VERSION_ID>`
   - `asc submit cancel --id <SUBMISSION_ID> --confirm`

## Notes
- Always use `--help` to verify flags for the exact command.
- Use `--output table` / `--output markdown` for human-readable output; default is JSON.
