package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

const (
	publishDefaultPollInterval = 30 * time.Second
	publishDefaultTimeout      = 30 * time.Minute
)

// PublishCommand returns the publish command with subcommands.
func PublishCommand() *ffcli.Command {
	return &ffcli.Command{
		Name:       "publish",
		ShortUsage: "asc publish <subcommand> [flags]",
		ShortHelp:  "End-to-end publish workflows for TestFlight and App Store.",
		LongHelp: `End-to-end publish workflows.

Combines upload, distribution, and submission into single commands.

Examples:
  asc publish testflight --app APP_ID --ipa app.ipa --group GROUP_ID
  asc publish appstore --app APP_ID --ipa app.ipa --version 1.2.3 --submit --confirm`,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			PublishTestFlightCommand(),
			PublishAppStoreCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// PublishTestFlightCommand uploads an IPA and distributes it to TestFlight groups.
func PublishTestFlightCommand() *ffcli.Command {
	fs := flag.NewFlagSet("publish testflight", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (required, or ASC_APP_ID env)")
	ipaPath := fs.String("ipa", "", "Path to .ipa file (required)")
	version := fs.String("version", "", "CFBundleShortVersionString (auto-extracted from IPA if not provided)")
	buildNumber := fs.String("build-number", "", "CFBundleVersion (auto-extracted from IPA if not provided)")
	platform := fs.String("platform", "IOS", "Platform: IOS, MAC_OS, TV_OS, VISION_OS")
	groupIDs := fs.String("group", "", "Beta group ID(s), comma-separated")
	notify := fs.Bool("notify", false, "Notify testers after adding to groups")
	wait := fs.Bool("wait", false, "Wait for build processing to complete")
	pollInterval := fs.Duration("poll-interval", publishDefaultPollInterval, "Polling interval for --wait and build discovery")
	timeout := fs.Duration("timeout", 0, "Override upload + processing timeout (e.g., 30m)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "testflight",
		ShortUsage: "asc publish testflight [flags]",
		ShortHelp:  "Upload and distribute to TestFlight.",
		LongHelp: `Upload IPA and distribute to TestFlight beta groups.

Steps:
1. Upload IPA to App Store Connect
2. Wait for processing (if --wait)
3. Add build to specified beta groups
4. Optionally notify testers

Examples:
  asc publish testflight --app "123" --ipa app.ipa --group "GROUP_ID"
  asc publish testflight --app "123" --ipa app.ipa --group "G1,G2" --wait --notify`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" {
				fmt.Fprintf(os.Stderr, "Error: --app is required (or set ASC_APP_ID)\n\n")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*ipaPath) == "" {
				fmt.Fprintf(os.Stderr, "Error: --ipa is required\n\n")
				return flag.ErrHelp
			}

			parsedGroupIDs := parseCommaSeparatedIDs(*groupIDs)
			if len(parsedGroupIDs) == 0 {
				fmt.Fprintf(os.Stderr, "Error: --group is required\n\n")
				return flag.ErrHelp
			}

			if *pollInterval <= 0 {
				return fmt.Errorf("publish testflight: --poll-interval must be greater than 0")
			}
			if *timeout < 0 {
				return fmt.Errorf("publish testflight: --timeout must be greater than 0")
			}

			normalizedPlatform, err := normalizeSubmitPlatform(*platform)
			if err != nil {
				return fmt.Errorf("publish testflight: %w", err)
			}

			fileInfo, err := validateIPAPath(*ipaPath)
			if err != nil {
				return fmt.Errorf("publish testflight: %w", err)
			}

			versionValue, buildNumberValue, err := resolveBundleInfoForIPA(*ipaPath, *version, *buildNumber)
			if err != nil {
				return fmt.Errorf("publish testflight: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("publish testflight: %w", err)
			}

			timeoutValue := resolvePublishTimeout(*timeout)
			requestCtx, cancel := contextWithPublishTimeout(ctx, timeoutValue)
			defer cancel()

			platformValue := asc.Platform(normalizedPlatform)
			timeoutOverride := *timeout > 0
			uploadResult, err := uploadBuildAndWaitForID(requestCtx, client, resolvedAppID, *ipaPath, fileInfo, versionValue, buildNumberValue, platformValue, *pollInterval, timeoutValue, timeoutOverride)
			if err != nil {
				return fmt.Errorf("publish testflight: %w", err)
			}

			buildResp := uploadResult.Build
			if *wait {
				buildResp, err = client.WaitForBuildProcessing(requestCtx, buildResp.Data.ID, *pollInterval)
				if err != nil {
					return fmt.Errorf("publish testflight: %w", err)
				}
			}

			if err := client.AddBetaGroupsToBuildWithNotify(requestCtx, buildResp.Data.ID, parsedGroupIDs, *notify); err != nil {
				return fmt.Errorf("publish testflight: failed to add groups: %w", err)
			}

			result := &asc.TestFlightPublishResult{
				BuildID:         buildResp.Data.ID,
				BuildVersion:    uploadResult.Version,
				BuildNumber:     uploadResult.BuildNumber,
				GroupIDs:        parsedGroupIDs,
				Uploaded:        true,
				ProcessingState: buildResp.Data.Attributes.ProcessingState,
				Notified:        *notify,
			}

			return printOutput(result, *output, *pretty)
		},
	}
}

// PublishAppStoreCommand uploads an IPA and submits it for App Store review.
func PublishAppStoreCommand() *ffcli.Command {
	fs := flag.NewFlagSet("publish appstore", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (required, or ASC_APP_ID env)")
	ipaPath := fs.String("ipa", "", "Path to .ipa file (required)")
	version := fs.String("version", "", "App Store version string (defaults to IPA version)")
	buildNumber := fs.String("build-number", "", "CFBundleVersion (auto-extracted from IPA if not provided)")
	platform := fs.String("platform", "IOS", "Platform: IOS, MAC_OS, TV_OS, VISION_OS")
	submit := fs.Bool("submit", false, "Submit for review after attaching build")
	confirm := fs.Bool("confirm", false, "Confirm submission (required with --submit)")
	wait := fs.Bool("wait", false, "Wait for build processing")
	pollInterval := fs.Duration("poll-interval", publishDefaultPollInterval, "Polling interval for --wait and build discovery")
	timeout := fs.Duration("timeout", 0, "Override upload + processing timeout (e.g., 30m)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "appstore",
		ShortUsage: "asc publish appstore [flags]",
		ShortHelp:  "Upload and submit to App Store.",
		LongHelp: `Upload IPA, attach to version, and optionally submit for review.

Steps:
1. Upload IPA to App Store Connect
2. Wait for processing (if --wait)
3. Find or create App Store version
4. Attach build to version
5. Submit for review (if --submit --confirm)

Examples:
  asc publish appstore --app "123" --ipa app.ipa --version 1.2.3
  asc publish appstore --app "123" --ipa app.ipa --version 1.2.3 --submit --confirm`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *submit && !*confirm {
				fmt.Fprintln(os.Stderr, "Error: --confirm is required with --submit")
				return flag.ErrHelp
			}

			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" {
				fmt.Fprintf(os.Stderr, "Error: --app is required (or set ASC_APP_ID)\n\n")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*ipaPath) == "" {
				fmt.Fprintf(os.Stderr, "Error: --ipa is required\n\n")
				return flag.ErrHelp
			}
			if *pollInterval <= 0 {
				return fmt.Errorf("publish appstore: --poll-interval must be greater than 0")
			}
			if *timeout < 0 {
				return fmt.Errorf("publish appstore: --timeout must be greater than 0")
			}

			normalizedPlatform, err := normalizeSubmitPlatform(*platform)
			if err != nil {
				return fmt.Errorf("publish appstore: %w", err)
			}

			fileInfo, err := validateIPAPath(*ipaPath)
			if err != nil {
				return fmt.Errorf("publish appstore: %w", err)
			}

			versionValue, buildNumberValue, err := resolveBundleInfoForIPA(*ipaPath, *version, *buildNumber)
			if err != nil {
				return fmt.Errorf("publish appstore: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("publish appstore: %w", err)
			}

			timeoutValue := resolvePublishTimeout(*timeout)
			requestCtx, cancel := contextWithPublishTimeout(ctx, timeoutValue)
			defer cancel()

			platformValue := asc.Platform(normalizedPlatform)
			timeoutOverride := *timeout > 0
			uploadResult, err := uploadBuildAndWaitForID(requestCtx, client, resolvedAppID, *ipaPath, fileInfo, versionValue, buildNumberValue, platformValue, *pollInterval, timeoutValue, timeoutOverride)
			if err != nil {
				return fmt.Errorf("publish appstore: %w", err)
			}

			buildResp := uploadResult.Build
			if *wait {
				buildResp, err = client.WaitForBuildProcessing(requestCtx, buildResp.Data.ID, *pollInterval)
				if err != nil {
					return fmt.Errorf("publish appstore: %w", err)
				}
			}

			versionResp, err := client.FindOrCreateAppStoreVersion(requestCtx, resolvedAppID, uploadResult.Version, platformValue)
			if err != nil {
				return fmt.Errorf("publish appstore: %w", err)
			}

			if err := client.AttachBuildToVersion(requestCtx, versionResp.Data.ID, buildResp.Data.ID); err != nil {
				return fmt.Errorf("publish appstore: failed to attach build: %w", err)
			}

			result := &asc.AppStorePublishResult{
				BuildID:   buildResp.Data.ID,
				VersionID: versionResp.Data.ID,
				Uploaded:  true,
				Attached:  true,
				Submitted: false,
			}

			if *submit {
				submitReq := asc.AppStoreVersionSubmissionCreateRequest{
					Data: asc.AppStoreVersionSubmissionCreateData{
						Type: asc.ResourceTypeAppStoreVersionSubmissions,
						Relationships: &asc.AppStoreVersionSubmissionRelationships{
							AppStoreVersion: &asc.Relationship{
								Data: asc.ResourceData{Type: asc.ResourceTypeAppStoreVersions, ID: versionResp.Data.ID},
							},
						},
					},
				}
				submitResp, err := client.CreateAppStoreVersionSubmission(requestCtx, submitReq)
				if err != nil {
					return fmt.Errorf("publish appstore: failed to submit: %w", err)
				}
				result.SubmissionID = submitResp.Data.ID
				result.Submitted = true
			}

			return printOutput(result, *output, *pretty)
		},
	}
}

type publishUploadResult struct {
	Build       *asc.BuildResponse
	Version     string
	BuildNumber string
}

func uploadBuildAndWaitForID(ctx context.Context, client *asc.Client, appID, ipaPath string, fileInfo os.FileInfo, version, buildNumber string, platform asc.Platform, pollInterval time.Duration, uploadTimeout time.Duration, overrideUploadTimeout bool) (*publishUploadResult, error) {
	_, fileResp, err := prepareBuildUpload(ctx, client, appID, fileInfo, version, buildNumber, platform)
	if err != nil {
		return nil, err
	}

	if len(fileResp.Data.Attributes.UploadOperations) == 0 {
		return nil, fmt.Errorf("no upload operations returned")
	}

	uploadCtx, uploadCancel := contextWithPublishUploadTimeout(ctx, uploadTimeout, overrideUploadTimeout)
	err = asc.ExecuteUploadOperations(uploadCtx, ipaPath, fileResp.Data.Attributes.UploadOperations)
	uploadCancel()
	if err != nil {
		return nil, err
	}

	commitCtx, commitCancel := contextWithPublishUploadTimeout(ctx, uploadTimeout, overrideUploadTimeout)
	err = commitBuildUploadFile(commitCtx, client, fileResp.Data.ID, nil)
	commitCancel()
	if err != nil {
		return nil, err
	}

	buildResp, err := waitForBuildByNumber(ctx, client, appID, version, buildNumber, string(platform), pollInterval)
	if err != nil {
		return nil, err
	}

	return &publishUploadResult{
		Build:       buildResp,
		Version:     version,
		BuildNumber: buildNumber,
	}, nil
}

func resolvePublishTimeout(timeout time.Duration) time.Duration {
	if timeout > 0 {
		return timeout
	}
	return asc.ResolveTimeoutWithDefault(publishDefaultTimeout)
}

func contextWithPublishTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, timeout)
}

func contextWithPublishUploadTimeout(ctx context.Context, timeout time.Duration, override bool) (context.Context, context.CancelFunc) {
	if override {
		if ctx == nil {
			ctx = context.Background()
		}
		return context.WithTimeout(ctx, timeout)
	}
	return contextWithUploadTimeout(ctx)
}

func validateIPAPath(ipaPath string) (os.FileInfo, error) {
	fileInfo, err := os.Stat(ipaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat IPA: %w", err)
	}
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("--ipa must be a file")
	}
	return fileInfo, nil
}

func resolveBundleInfoForIPA(ipaPath, version, buildNumber string) (string, string, error) {
	versionValue := strings.TrimSpace(version)
	buildNumberValue := strings.TrimSpace(buildNumber)
	if versionValue == "" || buildNumberValue == "" {
		info, err := extractBundleInfoFromIPA(ipaPath)
		if err != nil {
			missingFlags := make([]string, 0, 2)
			if versionValue == "" {
				missingFlags = append(missingFlags, "--version")
			}
			if buildNumberValue == "" {
				missingFlags = append(missingFlags, "--build-number")
			}
			return "", "", fmt.Errorf("%s required (failed to extract from IPA: %w)", strings.Join(missingFlags, " and "), err)
		}
		if versionValue == "" {
			versionValue = info.Version
		}
		if buildNumberValue == "" {
			buildNumberValue = info.BuildNumber
		}
	}
	if versionValue == "" || buildNumberValue == "" {
		missingFields := make([]string, 0, 2)
		missingFlags := make([]string, 0, 2)
		if versionValue == "" {
			missingFields = append(missingFields, "CFBundleShortVersionString")
			missingFlags = append(missingFlags, "--version")
		}
		if buildNumberValue == "" {
			missingFields = append(missingFields, "CFBundleVersion")
			missingFlags = append(missingFlags, "--build-number")
		}
		return "", "", fmt.Errorf("Info.plist missing %s; provide %s", strings.Join(missingFields, " and "), strings.Join(missingFlags, " and "))
	}
	return versionValue, buildNumberValue, nil
}

func prepareBuildUpload(ctx context.Context, client *asc.Client, appID string, fileInfo os.FileInfo, version, buildNumber string, platform asc.Platform) (*asc.BuildUploadResponse, *asc.BuildUploadFileResponse, error) {
	uploadReq := asc.BuildUploadCreateRequest{
		Data: asc.BuildUploadCreateData{
			Type: asc.ResourceTypeBuildUploads,
			Attributes: asc.BuildUploadAttributes{
				CFBundleShortVersionString: version,
				CFBundleVersion:            buildNumber,
				Platform:                   platform,
			},
			Relationships: &asc.BuildUploadRelationships{
				App: &asc.Relationship{
					Data: asc.ResourceData{Type: asc.ResourceTypeApps, ID: appID},
				},
			},
		},
	}

	uploadResp, err := client.CreateBuildUpload(ctx, uploadReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create upload record: %w", err)
	}

	fileReq := asc.BuildUploadFileCreateRequest{
		Data: asc.BuildUploadFileCreateData{
			Type: asc.ResourceTypeBuildUploadFiles,
			Attributes: asc.BuildUploadFileAttributes{
				FileName:  fileInfo.Name(),
				FileSize:  fileInfo.Size(),
				UTI:       asc.UTIIPA,
				AssetType: asc.AssetTypeAsset,
			},
			Relationships: &asc.BuildUploadFileRelationships{
				BuildUpload: &asc.Relationship{
					Data: asc.ResourceData{Type: asc.ResourceTypeBuildUploads, ID: uploadResp.Data.ID},
				},
			},
		},
	}

	fileResp, err := client.CreateBuildUploadFile(ctx, fileReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create file reservation: %w", err)
	}

	return uploadResp, fileResp, nil
}

func commitBuildUploadFile(ctx context.Context, client *asc.Client, fileID string, checksums *asc.Checksums) error {
	uploaded := true
	attrs := &asc.BuildUploadFileUpdateAttributes{
		Uploaded:            &uploaded,
		SourceFileChecksums: checksums,
	}
	req := asc.BuildUploadFileUpdateRequest{
		Data: asc.BuildUploadFileUpdateData{
			Type:       asc.ResourceTypeBuildUploadFiles,
			ID:         fileID,
			Attributes: attrs,
		},
	}

	if _, err := client.UpdateBuildUploadFile(ctx, fileID, req); err != nil {
		return fmt.Errorf("commit upload file: %w", err)
	}
	return nil
}

func waitForBuildByNumber(ctx context.Context, client *asc.Client, appID, version, buildNumber, platform string, pollInterval time.Duration) (*asc.BuildResponse, error) {
	if pollInterval <= 0 {
		pollInterval = publishDefaultPollInterval
	}
	buildNumber = strings.TrimSpace(buildNumber)
	if buildNumber == "" {
		return nil, fmt.Errorf("build number is required to resolve build")
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		build, err := findBuildByNumber(ctx, client, appID, version, buildNumber, platform)
		if err != nil {
			return nil, err
		}
		if build != nil {
			return build, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

func findBuildByNumber(ctx context.Context, client *asc.Client, appID, version, buildNumber, platform string) (*asc.BuildResponse, error) {
	preReleaseResp, err := client.GetPreReleaseVersions(ctx, appID,
		asc.WithPreReleaseVersionsVersion(version),
		asc.WithPreReleaseVersionsPlatform(platform),
		asc.WithPreReleaseVersionsLimit(10),
	)
	if err != nil {
		return nil, err
	}
	if len(preReleaseResp.Data) == 0 {
		return nil, nil
	}
	if len(preReleaseResp.Data) > 1 {
		return nil, fmt.Errorf("multiple pre-release versions found for version %q and platform %q", version, platform)
	}

	preReleaseID := preReleaseResp.Data[0].ID
	buildsResp, err := client.GetBuilds(ctx, appID,
		asc.WithBuildsPreReleaseVersion(preReleaseID),
		asc.WithBuildsSort("-uploadedDate"),
		asc.WithBuildsLimit(200),
	)
	if err != nil {
		return nil, err
	}
	for _, build := range buildsResp.Data {
		if strings.TrimSpace(build.Attributes.Version) == buildNumber {
			return &asc.BuildResponse{Data: build}, nil
		}
	}
	return nil, nil
}
