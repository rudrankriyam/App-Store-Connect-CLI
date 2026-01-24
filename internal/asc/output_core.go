package asc

import (
	"encoding/json"
	"os"
)

// PrintJSON prints data as minified JSON (best for AI agents)
func PrintJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	return enc.Encode(data)
}

// PrintPrettyJSON prints data as indented JSON (best for debugging).
func PrintPrettyJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// PrintMarkdown prints data as Markdown table
func PrintMarkdown(data interface{}) error {
	switch v := data.(type) {
	case *FeedbackResponse:
		return printFeedbackMarkdown(v)
	case *CrashesResponse:
		return printCrashesMarkdown(v)
	case *ReviewsResponse:
		return printReviewsMarkdown(v)
	case *AppsResponse:
		return printAppsMarkdown(v)
	case *AppResponse:
		return printAppsMarkdown(&AppsResponse{Data: []Resource[AppAttributes]{v.Data}})
	case *BuildsResponse:
		return printBuildsMarkdown(v)
	case *AppStoreVersionsResponse:
		return printAppStoreVersionsMarkdown(v)
	case *PreReleaseVersionsResponse:
		return printPreReleaseVersionsMarkdown(v)
	case *BuildResponse:
		return printBuildsMarkdown(&BuildsResponse{Data: []Resource[BuildAttributes]{v.Data}})
	case *PreReleaseVersionResponse:
		return printPreReleaseVersionsMarkdown(&PreReleaseVersionsResponse{Data: []PreReleaseVersion{v.Data}})
	case *AppStoreVersionLocalizationsResponse:
		return printAppStoreVersionLocalizationsMarkdown(v)
	case *AppStoreVersionLocalizationResponse:
		return printAppStoreVersionLocalizationsMarkdown(&AppStoreVersionLocalizationsResponse{Data: []Resource[AppStoreVersionLocalizationAttributes]{v.Data}})
	case *AppInfoLocalizationsResponse:
		return printAppInfoLocalizationsMarkdown(v)
	case *AppScreenshotSetsResponse:
		return printAppScreenshotSetsMarkdown(v)
	case *AppScreenshotSetResponse:
		return printAppScreenshotSetsMarkdown(&AppScreenshotSetsResponse{Data: []Resource[AppScreenshotSetAttributes]{v.Data}})
	case *AppScreenshotsResponse:
		return printAppScreenshotsMarkdown(v)
	case *AppScreenshotResponse:
		return printAppScreenshotsMarkdown(&AppScreenshotsResponse{Data: []Resource[AppScreenshotAttributes]{v.Data}})
	case *AppPreviewSetsResponse:
		return printAppPreviewSetsMarkdown(v)
	case *AppPreviewSetResponse:
		return printAppPreviewSetsMarkdown(&AppPreviewSetsResponse{Data: []Resource[AppPreviewSetAttributes]{v.Data}})
	case *AppPreviewsResponse:
		return printAppPreviewsMarkdown(v)
	case *AppPreviewResponse:
		return printAppPreviewsMarkdown(&AppPreviewsResponse{Data: []Resource[AppPreviewAttributes]{v.Data}})
	case *BetaGroupsResponse:
		return printBetaGroupsMarkdown(v)
	case *BetaGroupResponse:
		return printBetaGroupsMarkdown(&BetaGroupsResponse{Data: []Resource[BetaGroupAttributes]{v.Data}})
	case *BetaTestersResponse:
		return printBetaTestersMarkdown(v)
	case *BetaTesterResponse:
		return printBetaTesterMarkdown(v)
	case *SandboxTestersResponse:
		return printSandboxTestersMarkdown(v)
	case *SandboxTesterResponse:
		return printSandboxTestersMarkdown(&SandboxTestersResponse{Data: []Resource[SandboxTesterAttributes]{v.Data}})
	case *LocalizationDownloadResult:
		return printLocalizationDownloadResultMarkdown(v)
	case *LocalizationUploadResult:
		return printLocalizationUploadResultMarkdown(v)
	case *BuildUploadResult:
		return printBuildUploadResultMarkdown(v)
	case *AppScreenshotListResult:
		return printAppScreenshotListResultMarkdown(v)
	case *AppPreviewListResult:
		return printAppPreviewListResultMarkdown(v)
	case *AppScreenshotUploadResult:
		return printAppScreenshotUploadResultMarkdown(v)
	case *AppPreviewUploadResult:
		return printAppPreviewUploadResultMarkdown(v)
	case *AssetDeleteResult:
		return printAssetDeleteResultMarkdown(v)
	case *SalesReportResult:
		return printSalesReportResultMarkdown(v)
	case *FinanceReportResult:
		return printFinanceReportResultMarkdown(v)
	case *FinanceRegionsResult:
		return printFinanceRegionsMarkdown(v)
	case *AnalyticsReportRequestResult:
		return printAnalyticsReportRequestResultMarkdown(v)
	case *AnalyticsReportRequestsResponse:
		return printAnalyticsReportRequestsMarkdown(v)
	case *AnalyticsReportRequestResponse:
		return printAnalyticsReportRequestsMarkdown(&AnalyticsReportRequestsResponse{Data: []AnalyticsReportRequestResource{v.Data}, Links: v.Links})
	case *AnalyticsReportDownloadResult:
		return printAnalyticsReportDownloadResultMarkdown(v)
	case *AnalyticsReportGetResult:
		return printAnalyticsReportGetResultMarkdown(v)
	case *AppStoreVersionSubmissionResult:
		return printAppStoreVersionSubmissionMarkdown(v)
	case *AppStoreVersionSubmissionCreateResult:
		return printAppStoreVersionSubmissionCreateMarkdown(v)
	case *AppStoreVersionSubmissionStatusResult:
		return printAppStoreVersionSubmissionStatusMarkdown(v)
	case *AppStoreVersionSubmissionCancelResult:
		return printAppStoreVersionSubmissionCancelMarkdown(v)
	case *AppStoreVersionDetailResult:
		return printAppStoreVersionDetailMarkdown(v)
	case *AppStoreVersionAttachBuildResult:
		return printAppStoreVersionAttachBuildMarkdown(v)
	case *AppStoreVersionPhasedReleaseResponse:
		return printAppStoreVersionPhasedReleaseMarkdown(v)
	case *AppStoreVersionPhasedReleaseDeleteResult:
		return printAppStoreVersionPhasedReleaseDeleteResultMarkdown(v)
	case *BuildBetaGroupsUpdateResult:
		return printBuildBetaGroupsUpdateMarkdown(v)
	case *BetaTesterDeleteResult:
		return printBetaTesterDeleteResultMarkdown(v)
	case *BetaTesterGroupsUpdateResult:
		return printBetaTesterGroupsUpdateResultMarkdown(v)
	case *AppStoreVersionLocalizationDeleteResult:
		return printAppStoreVersionLocalizationDeleteResultMarkdown(v)
	case *BetaTesterInvitationResult:
		return printBetaTesterInvitationResultMarkdown(v)
	case *SandboxTesterDeleteResult:
		return printSandboxTesterDeleteResultMarkdown(v)
	case *SandboxTesterClearHistoryResult:
		return printSandboxTesterClearHistoryResultMarkdown(v)
	case *XcodeCloudRunResult:
		return printXcodeCloudRunResultMarkdown(v)
	case *XcodeCloudStatusResult:
		return printXcodeCloudStatusResultMarkdown(v)
	case *CiProductsResponse:
		return printCiProductsMarkdown(v)
	case *CiWorkflowsResponse:
		return printCiWorkflowsMarkdown(v)
	case *CiBuildRunsResponse:
		return printCiBuildRunsMarkdown(v)
	case *CustomerReviewResponseResponse:
		return printCustomerReviewResponseMarkdown(v)
	case *CustomerReviewResponseDeleteResult:
		return printCustomerReviewResponseDeleteResultMarkdown(v)
	default:
		return PrintJSON(data)
	}
}

// PrintTable prints data as a formatted table
func PrintTable(data interface{}) error {
	switch v := data.(type) {
	case *FeedbackResponse:
		return printFeedbackTable(v)
	case *CrashesResponse:
		return printCrashesTable(v)
	case *ReviewsResponse:
		return printReviewsTable(v)
	case *AppsResponse:
		return printAppsTable(v)
	case *AppResponse:
		return printAppsTable(&AppsResponse{Data: []Resource[AppAttributes]{v.Data}})
	case *BuildsResponse:
		return printBuildsTable(v)
	case *AppStoreVersionsResponse:
		return printAppStoreVersionsTable(v)
	case *PreReleaseVersionsResponse:
		return printPreReleaseVersionsTable(v)
	case *BuildResponse:
		return printBuildsTable(&BuildsResponse{Data: []Resource[BuildAttributes]{v.Data}})
	case *PreReleaseVersionResponse:
		return printPreReleaseVersionsTable(&PreReleaseVersionsResponse{Data: []PreReleaseVersion{v.Data}})
	case *AppStoreVersionLocalizationsResponse:
		return printAppStoreVersionLocalizationsTable(v)
	case *AppStoreVersionLocalizationResponse:
		return printAppStoreVersionLocalizationsTable(&AppStoreVersionLocalizationsResponse{Data: []Resource[AppStoreVersionLocalizationAttributes]{v.Data}})
	case *AppInfoLocalizationsResponse:
		return printAppInfoLocalizationsTable(v)
	case *AppScreenshotSetsResponse:
		return printAppScreenshotSetsTable(v)
	case *AppScreenshotSetResponse:
		return printAppScreenshotSetsTable(&AppScreenshotSetsResponse{Data: []Resource[AppScreenshotSetAttributes]{v.Data}})
	case *AppScreenshotsResponse:
		return printAppScreenshotsTable(v)
	case *AppScreenshotResponse:
		return printAppScreenshotsTable(&AppScreenshotsResponse{Data: []Resource[AppScreenshotAttributes]{v.Data}})
	case *AppPreviewSetsResponse:
		return printAppPreviewSetsTable(v)
	case *AppPreviewSetResponse:
		return printAppPreviewSetsTable(&AppPreviewSetsResponse{Data: []Resource[AppPreviewSetAttributes]{v.Data}})
	case *AppPreviewsResponse:
		return printAppPreviewsTable(v)
	case *AppPreviewResponse:
		return printAppPreviewsTable(&AppPreviewsResponse{Data: []Resource[AppPreviewAttributes]{v.Data}})
	case *BetaGroupsResponse:
		return printBetaGroupsTable(v)
	case *BetaGroupResponse:
		return printBetaGroupsTable(&BetaGroupsResponse{Data: []Resource[BetaGroupAttributes]{v.Data}})
	case *BetaTestersResponse:
		return printBetaTestersTable(v)
	case *BetaTesterResponse:
		return printBetaTesterTable(v)
	case *SandboxTestersResponse:
		return printSandboxTestersTable(v)
	case *SandboxTesterResponse:
		return printSandboxTestersTable(&SandboxTestersResponse{Data: []Resource[SandboxTesterAttributes]{v.Data}})
	case *LocalizationDownloadResult:
		return printLocalizationDownloadResultTable(v)
	case *LocalizationUploadResult:
		return printLocalizationUploadResultTable(v)
	case *BuildUploadResult:
		return printBuildUploadResultTable(v)
	case *AppScreenshotListResult:
		return printAppScreenshotListResultTable(v)
	case *AppPreviewListResult:
		return printAppPreviewListResultTable(v)
	case *AppScreenshotUploadResult:
		return printAppScreenshotUploadResultTable(v)
	case *AppPreviewUploadResult:
		return printAppPreviewUploadResultTable(v)
	case *AssetDeleteResult:
		return printAssetDeleteResultTable(v)
	case *SalesReportResult:
		return printSalesReportResultTable(v)
	case *FinanceReportResult:
		return printFinanceReportResultTable(v)
	case *FinanceRegionsResult:
		return printFinanceRegionsTable(v)
	case *AnalyticsReportRequestResult:
		return printAnalyticsReportRequestResultTable(v)
	case *AnalyticsReportRequestsResponse:
		return printAnalyticsReportRequestsTable(v)
	case *AnalyticsReportRequestResponse:
		return printAnalyticsReportRequestsTable(&AnalyticsReportRequestsResponse{Data: []AnalyticsReportRequestResource{v.Data}, Links: v.Links})
	case *AnalyticsReportDownloadResult:
		return printAnalyticsReportDownloadResultTable(v)
	case *AnalyticsReportGetResult:
		return printAnalyticsReportGetResultTable(v)
	case *AppStoreVersionSubmissionResult:
		return printAppStoreVersionSubmissionTable(v)
	case *AppStoreVersionSubmissionCreateResult:
		return printAppStoreVersionSubmissionCreateTable(v)
	case *AppStoreVersionSubmissionStatusResult:
		return printAppStoreVersionSubmissionStatusTable(v)
	case *AppStoreVersionSubmissionCancelResult:
		return printAppStoreVersionSubmissionCancelTable(v)
	case *AppStoreVersionDetailResult:
		return printAppStoreVersionDetailTable(v)
	case *AppStoreVersionAttachBuildResult:
		return printAppStoreVersionAttachBuildTable(v)
	case *AppStoreVersionPhasedReleaseResponse:
		return printAppStoreVersionPhasedReleaseTable(v)
	case *AppStoreVersionPhasedReleaseDeleteResult:
		return printAppStoreVersionPhasedReleaseDeleteResultTable(v)
	case *BuildBetaGroupsUpdateResult:
		return printBuildBetaGroupsUpdateTable(v)
	case *BetaTesterDeleteResult:
		return printBetaTesterDeleteResultTable(v)
	case *BetaTesterGroupsUpdateResult:
		return printBetaTesterGroupsUpdateResultTable(v)
	case *AppStoreVersionLocalizationDeleteResult:
		return printAppStoreVersionLocalizationDeleteResultTable(v)
	case *BetaTesterInvitationResult:
		return printBetaTesterInvitationResultTable(v)
	case *SandboxTesterDeleteResult:
		return printSandboxTesterDeleteResultTable(v)
	case *SandboxTesterClearHistoryResult:
		return printSandboxTesterClearHistoryResultTable(v)
	case *XcodeCloudRunResult:
		return printXcodeCloudRunResultTable(v)
	case *XcodeCloudStatusResult:
		return printXcodeCloudStatusResultTable(v)
	case *CiProductsResponse:
		return printCiProductsTable(v)
	case *CiWorkflowsResponse:
		return printCiWorkflowsTable(v)
	case *CiBuildRunsResponse:
		return printCiBuildRunsTable(v)
	case *CustomerReviewResponseResponse:
		return printCustomerReviewResponseTable(v)
	case *CustomerReviewResponseDeleteResult:
		return printCustomerReviewResponseDeleteResultTable(v)
	default:
		return PrintJSON(data)
	}
}
