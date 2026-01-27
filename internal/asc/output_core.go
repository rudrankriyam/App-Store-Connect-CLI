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
	case *AppCategoriesResponse:
		return printAppCategoriesMarkdown(v)
	case *AppResponse:
		return printAppsMarkdown(&AppsResponse{Data: []Resource[AppAttributes]{v.Data}})
	case *AppTagsResponse:
		return printAppTagsMarkdown(v)
	case *AppTagResponse:
		return printAppTagsMarkdown(&AppTagsResponse{Data: []Resource[AppTagAttributes]{v.Data}})
	case *LinkagesResponse:
		return printLinkagesMarkdown(v)
	case *BundleIDsResponse:
		return printBundleIDsMarkdown(v)
	case *BundleIDResponse:
		return printBundleIDsMarkdown(&BundleIDsResponse{Data: []Resource[BundleIDAttributes]{v.Data}})
	case *CertificatesResponse:
		return printCertificatesMarkdown(v)
	case *CertificateResponse:
		return printCertificatesMarkdown(&CertificatesResponse{Data: []Resource[CertificateAttributes]{v.Data}})
	case *ProfilesResponse:
		return printProfilesMarkdown(v)
	case *ProfileResponse:
		return printProfilesMarkdown(&ProfilesResponse{Data: []Resource[ProfileAttributes]{v.Data}})
	case *InAppPurchasesV2Response:
		return printInAppPurchasesMarkdown(v)
	case *InAppPurchaseV2Response:
		return printInAppPurchasesMarkdown(&InAppPurchasesV2Response{Data: []Resource[InAppPurchaseV2Attributes]{v.Data}})
	case *InAppPurchaseLocalizationsResponse:
		return printInAppPurchaseLocalizationsMarkdown(v)
	case *SubscriptionGroupsResponse:
		return printSubscriptionGroupsMarkdown(v)
	case *SubscriptionGroupResponse:
		return printSubscriptionGroupsMarkdown(&SubscriptionGroupsResponse{Data: []Resource[SubscriptionGroupAttributes]{v.Data}})
	case *SubscriptionsResponse:
		return printSubscriptionsMarkdown(v)
	case *SubscriptionResponse:
		return printSubscriptionsMarkdown(&SubscriptionsResponse{Data: []Resource[SubscriptionAttributes]{v.Data}})
	case *SubscriptionPriceResponse:
		return printSubscriptionPriceMarkdown(v)
	case *SubscriptionAvailabilityResponse:
		return printSubscriptionAvailabilityMarkdown(v)
	case *TerritoriesResponse:
		return printTerritoriesMarkdown(v)
	case *AppPricePointsV3Response:
		return printAppPricePointsMarkdown(v)
	case *AppPriceScheduleResponse:
		return printAppPriceScheduleMarkdown(v)
	case *AppPricesResponse:
		return printAppPricesMarkdown(v)
	case *BuildsResponse:
		return printBuildsMarkdown(v)
	case *BuildBundlesResponse:
		return printBuildBundlesMarkdown(v)
	case *BuildBundleFileSizesResponse:
		return printBuildBundleFileSizesMarkdown(v)
	case *BetaAppClipInvocationsResponse:
		return printBetaAppClipInvocationsMarkdown(v)
	case *SubscriptionOfferCodeOneTimeUseCodesResponse:
		return printOfferCodesMarkdown(v)
	case *AppStoreVersionsResponse:
		return printAppStoreVersionsMarkdown(v)
	case *PreReleaseVersionsResponse:
		return printPreReleaseVersionsMarkdown(v)
	case *BuildResponse:
		return printBuildsMarkdown(&BuildsResponse{Data: []Resource[BuildAttributes]{v.Data}})
	case *AppClipDomainStatusResult:
		return printAppClipDomainStatusResultMarkdown(v)
	case *SubscriptionOfferCodeOneTimeUseCodeResponse:
		return printOfferCodesMarkdown(&SubscriptionOfferCodeOneTimeUseCodesResponse{Data: []Resource[SubscriptionOfferCodeOneTimeUseCodeAttributes]{v.Data}})
	case *AppAvailabilityV2Response:
		return printAppAvailabilityMarkdown(v)
	case *TerritoryAvailabilitiesResponse:
		return printTerritoryAvailabilitiesMarkdown(v)
	case *PreReleaseVersionResponse:
		return printPreReleaseVersionsMarkdown(&PreReleaseVersionsResponse{Data: []PreReleaseVersion{v.Data}})
	case *AppStoreVersionLocalizationsResponse:
		return printAppStoreVersionLocalizationsMarkdown(v)
	case *AppStoreVersionLocalizationResponse:
		return printAppStoreVersionLocalizationsMarkdown(&AppStoreVersionLocalizationsResponse{Data: []Resource[AppStoreVersionLocalizationAttributes]{v.Data}})
	case *BetaBuildLocalizationsResponse:
		return printBetaBuildLocalizationsMarkdown(v)
	case *BetaBuildLocalizationResponse:
		return printBetaBuildLocalizationsMarkdown(&BetaBuildLocalizationsResponse{Data: []Resource[BetaBuildLocalizationAttributes]{v.Data}})
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
	case *UsersResponse:
		return printUsersMarkdown(v)
	case *UserResponse:
		return printUsersMarkdown(&UsersResponse{Data: []Resource[UserAttributes]{v.Data}})
	case *DevicesResponse:
		return printDevicesMarkdown(v)
	case *DeviceLocalUDIDResult:
		return printDeviceLocalUDIDMarkdown(v)
	case *DeviceResponse:
		return printDevicesMarkdown(&DevicesResponse{Data: []Resource[DeviceAttributes]{v.Data}})
	case *UserInvitationsResponse:
		return printUserInvitationsMarkdown(v)
	case *UserInvitationResponse:
		return printUserInvitationsMarkdown(&UserInvitationsResponse{Data: []Resource[UserInvitationAttributes]{v.Data}})
	case *UserDeleteResult:
		return printUserDeleteResultMarkdown(v)
	case *UserInvitationRevokeResult:
		return printUserInvitationRevokeResultMarkdown(v)
	case *BetaAppReviewDetailsResponse:
		return printBetaAppReviewDetailsMarkdown(v)
	case *BetaAppReviewDetailResponse:
		return printBetaAppReviewDetailMarkdown(v)
	case *BetaAppReviewSubmissionsResponse:
		return printBetaAppReviewSubmissionsMarkdown(v)
	case *BetaAppReviewSubmissionResponse:
		return printBetaAppReviewSubmissionMarkdown(v)
	case *BuildBetaDetailsResponse:
		return printBuildBetaDetailsMarkdown(v)
	case *BuildBetaDetailResponse:
		return printBuildBetaDetailMarkdown(v)
	case *AgeRatingDeclarationResponse:
		return printAgeRatingDeclarationMarkdown(v)
	case *AccessibilityDeclarationsResponse:
		return printAccessibilityDeclarationsMarkdown(v)
	case *AccessibilityDeclarationResponse:
		return printAccessibilityDeclarationMarkdown(v)
	case *AppStoreReviewDetailResponse:
		return printAppStoreReviewDetailMarkdown(v)
	case *AppStoreReviewAttachmentsResponse:
		return printAppStoreReviewAttachmentsMarkdown(v)
	case *AppStoreReviewAttachmentResponse:
		return printAppStoreReviewAttachmentMarkdown(v)
	case *RoutingAppCoverageResponse:
		return printRoutingAppCoverageMarkdown(v)
	case *AppEncryptionDeclarationsResponse:
		return printAppEncryptionDeclarationsMarkdown(v)
	case *AppEncryptionDeclarationResponse:
		return printAppEncryptionDeclarationMarkdown(v)
	case *AppEncryptionDeclarationDocumentResponse:
		return printAppEncryptionDeclarationDocumentMarkdown(v)
	case *BetaRecruitmentCriterionOptionsResponse:
		return printBetaRecruitmentCriterionOptionsMarkdown(v)
	case *BetaRecruitmentCriteriaResponse:
		return printBetaRecruitmentCriteriaMarkdown(v)
	case *Response[BetaGroupMetricAttributes]:
		return printBetaGroupMetricsMarkdown(v.Data)
	case *SandboxTestersResponse:
		return printSandboxTestersMarkdown(v)
	case *SandboxTesterResponse:
		return printSandboxTestersMarkdown(&SandboxTestersResponse{Data: []Resource[SandboxTesterAttributes]{v.Data}})
	case *BundleIDCapabilitiesResponse:
		return printBundleIDCapabilitiesMarkdown(v)
	case *BundleIDCapabilityResponse:
		return printBundleIDCapabilitiesMarkdown(&BundleIDCapabilitiesResponse{Data: []Resource[BundleIDCapabilityAttributes]{v.Data}})
	case *LocalizationDownloadResult:
		return printLocalizationDownloadResultMarkdown(v)
	case *LocalizationUploadResult:
		return printLocalizationUploadResultMarkdown(v)
	case *BuildUploadResult:
		return printBuildUploadResultMarkdown(v)
	case *BuildExpireAllResult:
		return printBuildExpireAllResultMarkdown(v)
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
	case *TestFlightPublishResult:
		return printTestFlightPublishResultMarkdown(v)
	case *AppStorePublishResult:
		return printAppStorePublishResultMarkdown(v)
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
	case *ReviewSubmissionsResponse:
		return printReviewSubmissionsMarkdown(v)
	case *ReviewSubmissionResponse:
		return printReviewSubmissionsMarkdown(&ReviewSubmissionsResponse{Data: []ReviewSubmissionResource{v.Data}, Links: v.Links})
	case *ReviewSubmissionItemsResponse:
		return printReviewSubmissionItemsMarkdown(v)
	case *ReviewSubmissionItemResponse:
		return printReviewSubmissionItemsMarkdown(&ReviewSubmissionItemsResponse{Data: []ReviewSubmissionItemResource{v.Data}, Links: v.Links})
	case *ReviewSubmissionItemDeleteResult:
		return printReviewSubmissionItemDeleteResultMarkdown(v)
	case *AppStoreVersionReleaseRequestResult:
		return printAppStoreVersionReleaseRequestMarkdown(v)
	case *AppStoreVersionPromotionCreateResult:
		return printAppStoreVersionPromotionCreateMarkdown(v)
	case *AppStoreVersionPhasedReleaseResponse:
		return printAppStoreVersionPhasedReleaseMarkdown(v)
	case *AppStoreVersionPhasedReleaseDeleteResult:
		return printAppStoreVersionPhasedReleaseDeleteResultMarkdown(v)
	case *BuildBetaGroupsUpdateResult:
		return printBuildBetaGroupsUpdateMarkdown(v)
	case *InAppPurchaseDeleteResult:
		return printInAppPurchaseDeleteResultMarkdown(v)
	case *SubscriptionGroupDeleteResult:
		return printSubscriptionGroupDeleteResultMarkdown(v)
	case *SubscriptionDeleteResult:
		return printSubscriptionDeleteResultMarkdown(v)
	case *BetaTesterDeleteResult:
		return printBetaTesterDeleteResultMarkdown(v)
	case *BetaTesterGroupsUpdateResult:
		return printBetaTesterGroupsUpdateResultMarkdown(v)
	case *AppStoreVersionLocalizationDeleteResult:
		return printAppStoreVersionLocalizationDeleteResultMarkdown(v)
	case *BetaBuildLocalizationDeleteResult:
		return printBetaBuildLocalizationDeleteResultMarkdown(v)
	case *BetaTesterInvitationResult:
		return printBetaTesterInvitationResultMarkdown(v)
	case *SandboxTesterDeleteResult:
		return printSandboxTesterDeleteResultMarkdown(v)
	case *SandboxTesterClearHistoryResult:
		return printSandboxTesterClearHistoryResultMarkdown(v)
	case *BundleIDDeleteResult:
		return printBundleIDDeleteResultMarkdown(v)
	case *BundleIDCapabilityDeleteResult:
		return printBundleIDCapabilityDeleteResultMarkdown(v)
	case *CertificateRevokeResult:
		return printCertificateRevokeResultMarkdown(v)
	case *ProfileDeleteResult:
		return printProfileDeleteResultMarkdown(v)
	case *ProfileDownloadResult:
		return printProfileDownloadResultMarkdown(v)
	case *SigningFetchResult:
		return printSigningFetchResultMarkdown(v)
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
	case *CiBuildActionsResponse:
		return printCiBuildActionsMarkdown(v)
	case *CiArtifactsResponse:
		return printCiArtifactsMarkdown(v)
	case *CiArtifactResponse:
		return printCiArtifactMarkdown(v)
	case *CiTestResultsResponse:
		return printCiTestResultsMarkdown(v)
	case *CiTestResultResponse:
		return printCiTestResultMarkdown(v)
	case *CiIssuesResponse:
		return printCiIssuesMarkdown(v)
	case *CiIssueResponse:
		return printCiIssueMarkdown(v)
	case *CiArtifactDownloadResult:
		return printCiArtifactDownloadResultMarkdown(v)
	case *EndUserLicenseAgreementResponse:
		return printEndUserLicenseAgreementMarkdown(v)
	case *EndUserLicenseAgreementDeleteResult:
		return printEndUserLicenseAgreementDeleteResultMarkdown(v)
	case *CustomerReviewResponseResponse:
		return printCustomerReviewResponseMarkdown(v)
	case *CustomerReviewResponseDeleteResult:
		return printCustomerReviewResponseDeleteResultMarkdown(v)
	case *AccessibilityDeclarationDeleteResult:
		return printAccessibilityDeclarationDeleteResultMarkdown(v)
	case *AppStoreReviewAttachmentDeleteResult:
		return printAppStoreReviewAttachmentDeleteResultMarkdown(v)
	case *RoutingAppCoverageDeleteResult:
		return printRoutingAppCoverageDeleteResultMarkdown(v)
	case *AppEncryptionDeclarationBuildsUpdateResult:
		return printAppEncryptionDeclarationBuildsUpdateResultMarkdown(v)
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
	case *AppCategoriesResponse:
		return printAppCategoriesTable(v)
	case *AppResponse:
		return printAppsTable(&AppsResponse{Data: []Resource[AppAttributes]{v.Data}})
	case *AppTagsResponse:
		return printAppTagsTable(v)
	case *AppTagResponse:
		return printAppTagsTable(&AppTagsResponse{Data: []Resource[AppTagAttributes]{v.Data}})
	case *LinkagesResponse:
		return printLinkagesTable(v)
	case *BundleIDsResponse:
		return printBundleIDsTable(v)
	case *BundleIDResponse:
		return printBundleIDsTable(&BundleIDsResponse{Data: []Resource[BundleIDAttributes]{v.Data}})
	case *CertificatesResponse:
		return printCertificatesTable(v)
	case *CertificateResponse:
		return printCertificatesTable(&CertificatesResponse{Data: []Resource[CertificateAttributes]{v.Data}})
	case *ProfilesResponse:
		return printProfilesTable(v)
	case *ProfileResponse:
		return printProfilesTable(&ProfilesResponse{Data: []Resource[ProfileAttributes]{v.Data}})
	case *InAppPurchasesV2Response:
		return printInAppPurchasesTable(v)
	case *InAppPurchaseV2Response:
		return printInAppPurchasesTable(&InAppPurchasesV2Response{Data: []Resource[InAppPurchaseV2Attributes]{v.Data}})
	case *InAppPurchaseLocalizationsResponse:
		return printInAppPurchaseLocalizationsTable(v)
	case *SubscriptionGroupsResponse:
		return printSubscriptionGroupsTable(v)
	case *SubscriptionGroupResponse:
		return printSubscriptionGroupsTable(&SubscriptionGroupsResponse{Data: []Resource[SubscriptionGroupAttributes]{v.Data}})
	case *SubscriptionsResponse:
		return printSubscriptionsTable(v)
	case *SubscriptionResponse:
		return printSubscriptionsTable(&SubscriptionsResponse{Data: []Resource[SubscriptionAttributes]{v.Data}})
	case *SubscriptionPriceResponse:
		return printSubscriptionPriceTable(v)
	case *SubscriptionAvailabilityResponse:
		return printSubscriptionAvailabilityTable(v)
	case *TerritoriesResponse:
		return printTerritoriesTable(v)
	case *AppPricePointsV3Response:
		return printAppPricePointsTable(v)
	case *AppPriceScheduleResponse:
		return printAppPriceScheduleTable(v)
	case *AppPricesResponse:
		return printAppPricesTable(v)
	case *BuildsResponse:
		return printBuildsTable(v)
	case *BuildBundlesResponse:
		return printBuildBundlesTable(v)
	case *BuildBundleFileSizesResponse:
		return printBuildBundleFileSizesTable(v)
	case *BetaAppClipInvocationsResponse:
		return printBetaAppClipInvocationsTable(v)
	case *SubscriptionOfferCodeOneTimeUseCodesResponse:
		return printOfferCodesTable(v)
	case *AppStoreVersionsResponse:
		return printAppStoreVersionsTable(v)
	case *PreReleaseVersionsResponse:
		return printPreReleaseVersionsTable(v)
	case *BuildResponse:
		return printBuildsTable(&BuildsResponse{Data: []Resource[BuildAttributes]{v.Data}})
	case *AppClipDomainStatusResult:
		return printAppClipDomainStatusResultTable(v)
	case *SubscriptionOfferCodeOneTimeUseCodeResponse:
		return printOfferCodesTable(&SubscriptionOfferCodeOneTimeUseCodesResponse{Data: []Resource[SubscriptionOfferCodeOneTimeUseCodeAttributes]{v.Data}})
	case *AppAvailabilityV2Response:
		return printAppAvailabilityTable(v)
	case *TerritoryAvailabilitiesResponse:
		return printTerritoryAvailabilitiesTable(v)
	case *PreReleaseVersionResponse:
		return printPreReleaseVersionsTable(&PreReleaseVersionsResponse{Data: []PreReleaseVersion{v.Data}})
	case *AppStoreVersionLocalizationsResponse:
		return printAppStoreVersionLocalizationsTable(v)
	case *AppStoreVersionLocalizationResponse:
		return printAppStoreVersionLocalizationsTable(&AppStoreVersionLocalizationsResponse{Data: []Resource[AppStoreVersionLocalizationAttributes]{v.Data}})
	case *BetaBuildLocalizationsResponse:
		return printBetaBuildLocalizationsTable(v)
	case *BetaBuildLocalizationResponse:
		return printBetaBuildLocalizationsTable(&BetaBuildLocalizationsResponse{Data: []Resource[BetaBuildLocalizationAttributes]{v.Data}})
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
	case *UsersResponse:
		return printUsersTable(v)
	case *UserResponse:
		return printUsersTable(&UsersResponse{Data: []Resource[UserAttributes]{v.Data}})
	case *DevicesResponse:
		return printDevicesTable(v)
	case *DeviceLocalUDIDResult:
		return printDeviceLocalUDIDTable(v)
	case *DeviceResponse:
		return printDevicesTable(&DevicesResponse{Data: []Resource[DeviceAttributes]{v.Data}})
	case *UserInvitationsResponse:
		return printUserInvitationsTable(v)
	case *UserInvitationResponse:
		return printUserInvitationsTable(&UserInvitationsResponse{Data: []Resource[UserInvitationAttributes]{v.Data}})
	case *UserDeleteResult:
		return printUserDeleteResultTable(v)
	case *UserInvitationRevokeResult:
		return printUserInvitationRevokeResultTable(v)
	case *BetaAppReviewDetailsResponse:
		return printBetaAppReviewDetailsTable(v)
	case *BetaAppReviewDetailResponse:
		return printBetaAppReviewDetailTable(v)
	case *BetaAppReviewSubmissionsResponse:
		return printBetaAppReviewSubmissionsTable(v)
	case *BetaAppReviewSubmissionResponse:
		return printBetaAppReviewSubmissionTable(v)
	case *BuildBetaDetailsResponse:
		return printBuildBetaDetailsTable(v)
	case *BuildBetaDetailResponse:
		return printBuildBetaDetailTable(v)
	case *AgeRatingDeclarationResponse:
		return printAgeRatingDeclarationTable(v)
	case *AccessibilityDeclarationsResponse:
		return printAccessibilityDeclarationsTable(v)
	case *AccessibilityDeclarationResponse:
		return printAccessibilityDeclarationTable(v)
	case *AppStoreReviewDetailResponse:
		return printAppStoreReviewDetailTable(v)
	case *AppStoreReviewAttachmentsResponse:
		return printAppStoreReviewAttachmentsTable(v)
	case *AppStoreReviewAttachmentResponse:
		return printAppStoreReviewAttachmentTable(v)
	case *RoutingAppCoverageResponse:
		return printRoutingAppCoverageTable(v)
	case *AppEncryptionDeclarationsResponse:
		return printAppEncryptionDeclarationsTable(v)
	case *AppEncryptionDeclarationResponse:
		return printAppEncryptionDeclarationTable(v)
	case *AppEncryptionDeclarationDocumentResponse:
		return printAppEncryptionDeclarationDocumentTable(v)
	case *BetaRecruitmentCriterionOptionsResponse:
		return printBetaRecruitmentCriterionOptionsTable(v)
	case *BetaRecruitmentCriteriaResponse:
		return printBetaRecruitmentCriteriaTable(v)
	case *Response[BetaGroupMetricAttributes]:
		return printBetaGroupMetricsTable(v.Data)
	case *SandboxTestersResponse:
		return printSandboxTestersTable(v)
	case *SandboxTesterResponse:
		return printSandboxTestersTable(&SandboxTestersResponse{Data: []Resource[SandboxTesterAttributes]{v.Data}})
	case *BundleIDCapabilitiesResponse:
		return printBundleIDCapabilitiesTable(v)
	case *BundleIDCapabilityResponse:
		return printBundleIDCapabilitiesTable(&BundleIDCapabilitiesResponse{Data: []Resource[BundleIDCapabilityAttributes]{v.Data}})
	case *LocalizationDownloadResult:
		return printLocalizationDownloadResultTable(v)
	case *LocalizationUploadResult:
		return printLocalizationUploadResultTable(v)
	case *BuildUploadResult:
		return printBuildUploadResultTable(v)
	case *BuildExpireAllResult:
		return printBuildExpireAllResultTable(v)
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
	case *TestFlightPublishResult:
		return printTestFlightPublishResultTable(v)
	case *AppStorePublishResult:
		return printAppStorePublishResultTable(v)
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
	case *ReviewSubmissionsResponse:
		return printReviewSubmissionsTable(v)
	case *ReviewSubmissionResponse:
		return printReviewSubmissionsTable(&ReviewSubmissionsResponse{Data: []ReviewSubmissionResource{v.Data}, Links: v.Links})
	case *ReviewSubmissionItemsResponse:
		return printReviewSubmissionItemsTable(v)
	case *ReviewSubmissionItemResponse:
		return printReviewSubmissionItemsTable(&ReviewSubmissionItemsResponse{Data: []ReviewSubmissionItemResource{v.Data}, Links: v.Links})
	case *ReviewSubmissionItemDeleteResult:
		return printReviewSubmissionItemDeleteResultTable(v)
	case *AppStoreVersionReleaseRequestResult:
		return printAppStoreVersionReleaseRequestTable(v)
	case *AppStoreVersionPromotionCreateResult:
		return printAppStoreVersionPromotionCreateTable(v)
	case *AppStoreVersionPhasedReleaseResponse:
		return printAppStoreVersionPhasedReleaseTable(v)
	case *AppStoreVersionPhasedReleaseDeleteResult:
		return printAppStoreVersionPhasedReleaseDeleteResultTable(v)
	case *BuildBetaGroupsUpdateResult:
		return printBuildBetaGroupsUpdateTable(v)
	case *InAppPurchaseDeleteResult:
		return printInAppPurchaseDeleteResultTable(v)
	case *SubscriptionGroupDeleteResult:
		return printSubscriptionGroupDeleteResultTable(v)
	case *SubscriptionDeleteResult:
		return printSubscriptionDeleteResultTable(v)
	case *BetaTesterDeleteResult:
		return printBetaTesterDeleteResultTable(v)
	case *BetaTesterGroupsUpdateResult:
		return printBetaTesterGroupsUpdateResultTable(v)
	case *AppStoreVersionLocalizationDeleteResult:
		return printAppStoreVersionLocalizationDeleteResultTable(v)
	case *BetaBuildLocalizationDeleteResult:
		return printBetaBuildLocalizationDeleteResultTable(v)
	case *BetaTesterInvitationResult:
		return printBetaTesterInvitationResultTable(v)
	case *SandboxTesterDeleteResult:
		return printSandboxTesterDeleteResultTable(v)
	case *SandboxTesterClearHistoryResult:
		return printSandboxTesterClearHistoryResultTable(v)
	case *BundleIDDeleteResult:
		return printBundleIDDeleteResultTable(v)
	case *BundleIDCapabilityDeleteResult:
		return printBundleIDCapabilityDeleteResultTable(v)
	case *CertificateRevokeResult:
		return printCertificateRevokeResultTable(v)
	case *ProfileDeleteResult:
		return printProfileDeleteResultTable(v)
	case *EndUserLicenseAgreementResponse:
		return printEndUserLicenseAgreementTable(v)
	case *EndUserLicenseAgreementDeleteResult:
		return printEndUserLicenseAgreementDeleteResultTable(v)
	case *ProfileDownloadResult:
		return printProfileDownloadResultTable(v)
	case *SigningFetchResult:
		return printSigningFetchResultTable(v)
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
	case *CiBuildActionsResponse:
		return printCiBuildActionsTable(v)
	case *CiArtifactsResponse:
		return printCiArtifactsTable(v)
	case *CiArtifactResponse:
		return printCiArtifactTable(v)
	case *CiTestResultsResponse:
		return printCiTestResultsTable(v)
	case *CiTestResultResponse:
		return printCiTestResultTable(v)
	case *CiIssuesResponse:
		return printCiIssuesTable(v)
	case *CiIssueResponse:
		return printCiIssueTable(v)
	case *CiArtifactDownloadResult:
		return printCiArtifactDownloadResultTable(v)
	case *CustomerReviewResponseResponse:
		return printCustomerReviewResponseTable(v)
	case *CustomerReviewResponseDeleteResult:
		return printCustomerReviewResponseDeleteResultTable(v)
	case *AccessibilityDeclarationDeleteResult:
		return printAccessibilityDeclarationDeleteResultTable(v)
	case *AppStoreReviewAttachmentDeleteResult:
		return printAppStoreReviewAttachmentDeleteResultTable(v)
	case *RoutingAppCoverageDeleteResult:
		return printRoutingAppCoverageDeleteResultTable(v)
	case *AppEncryptionDeclarationBuildsUpdateResult:
		return printAppEncryptionDeclarationBuildsUpdateResultTable(v)
	default:
		return PrintJSON(data)
	}
}
