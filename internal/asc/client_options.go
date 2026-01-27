package asc

import "strings"

// FeedbackOption is a functional option for GetFeedback.
type FeedbackOption func(*feedbackQuery)

// CrashOption is a functional option for GetCrashes.
type CrashOption func(*crashQuery)

// ReviewOption is a functional option for GetReviews.
type ReviewOption func(*reviewQuery)

// AppsOption is a functional option for GetApps.
type AppsOption func(*appsQuery)

// AppTagsOption is a functional option for GetAppTags.
type AppTagsOption func(*appTagsQuery)

// BuildsOption is a functional option for GetBuilds.
type BuildsOption func(*buildsQuery)

// BuildBundlesOption is a functional option for GetBuildBundlesForBuild.
type BuildBundlesOption func(*buildBundlesQuery)

// BuildBundleFileSizesOption is a functional option for GetBuildBundleFileSizes.
type BuildBundleFileSizesOption func(*buildBundleFileSizesQuery)

// BetaAppClipInvocationsOption is a functional option for GetBuildBundleBetaAppClipInvocations.
type BetaAppClipInvocationsOption func(*betaAppClipInvocationsQuery)

// SubscriptionOfferCodeOneTimeUseCodesOption is a functional option for GetSubscriptionOfferCodeOneTimeUseCodes.
type SubscriptionOfferCodeOneTimeUseCodesOption func(*subscriptionOfferCodeOneTimeUseCodesQuery)

// AppStoreVersionsOption is a functional option for GetAppStoreVersions.
type AppStoreVersionsOption func(*appStoreVersionsQuery)

// ReviewSubmissionsOption is a functional option for GetReviewSubmissions.
type ReviewSubmissionsOption func(*reviewSubmissionsQuery)

// ReviewSubmissionItemsOption is a functional option for GetReviewSubmissionItems.
type ReviewSubmissionItemsOption func(*reviewSubmissionItemsQuery)

// PreReleaseVersionsOption is a functional option for GetPreReleaseVersions.
type PreReleaseVersionsOption func(*preReleaseVersionsQuery)

// BetaGroupsOption is a functional option for GetBetaGroups.
type BetaGroupsOption func(*betaGroupsQuery)

// BetaGroupBuildsOption is a functional option for GetBetaGroupBuilds.
type BetaGroupBuildsOption func(*betaGroupBuildsQuery)

// BetaGroupTestersOption is a functional option for GetBetaGroupTesters.
type BetaGroupTestersOption func(*betaGroupTestersQuery)

// BetaTestersOption is a functional option for GetBetaTesters.
type BetaTestersOption func(*betaTestersQuery)

// BundleIDsOption is a functional option for GetBundleIDs.
type BundleIDsOption func(*bundleIDsQuery)

// BundleIDCapabilitiesOption is a functional option for GetBundleIDCapabilities.
type BundleIDCapabilitiesOption func(*bundleIDCapabilitiesQuery)

// CertificatesOption is a functional option for GetCertificates.
type CertificatesOption func(*certificatesQuery)

// DevicesOption is a functional option for GetDevices.
type DevicesOption func(*devicesQuery)

// ProfilesOption is a functional option for GetProfiles.
type ProfilesOption func(*profilesQuery)

// UsersOption is a functional option for GetUsers.
type UsersOption func(*usersQuery)

// UserInvitationsOption is a functional option for GetUserInvitations.
type UserInvitationsOption func(*userInvitationsQuery)

// BetaAppReviewDetailsOption is a functional option for beta app review details.
type BetaAppReviewDetailsOption func(*betaAppReviewDetailsQuery)

// BetaAppReviewSubmissionsOption is a functional option for beta app review submissions.
type BetaAppReviewSubmissionsOption func(*betaAppReviewSubmissionsQuery)

// BuildBetaDetailsOption is a functional option for build beta details.
type BuildBetaDetailsOption func(*buildBetaDetailsQuery)

// BetaRecruitmentCriterionOptionsOption is a functional option for recruitment options.
type BetaRecruitmentCriterionOptionsOption func(*betaRecruitmentCriterionOptionsQuery)

// AppStoreVersionLocalizationsOption is a functional option for version localizations.
type AppStoreVersionLocalizationsOption func(*appStoreVersionLocalizationsQuery)

// BetaBuildLocalizationsOption is a functional option for beta build localizations.
type BetaBuildLocalizationsOption func(*betaBuildLocalizationsQuery)

// AppInfoLocalizationsOption is a functional option for app info localizations.
type AppInfoLocalizationsOption func(*appInfoLocalizationsQuery)

// TerritoriesOption is a functional option for GetTerritories.
type TerritoriesOption func(*territoriesQuery)

// LinkagesOption is a functional option for linkages endpoints.
type LinkagesOption func(*linkagesQuery)

// PricePointsOption is a functional option for GetAppPricePoints.
type PricePointsOption func(*pricePointsQuery)

// AccessibilityDeclarationsOption is a functional option for accessibility declarations.
type AccessibilityDeclarationsOption func(*accessibilityDeclarationsQuery)

// AppStoreReviewAttachmentsOption is a functional option for review attachments.
type AppStoreReviewAttachmentsOption func(*appStoreReviewAttachmentsQuery)

// AppEncryptionDeclarationsOption is a functional option for encryption declarations.
type AppEncryptionDeclarationsOption func(*appEncryptionDeclarationsQuery)

// WithFeedbackDeviceModels filters feedback by device model(s).
func WithFeedbackDeviceModels(models []string) FeedbackOption {
	return func(q *feedbackQuery) {
		q.deviceModels = normalizeList(models)
	}
}

// WithFeedbackOSVersions filters feedback by OS version(s).
func WithFeedbackOSVersions(versions []string) FeedbackOption {
	return func(q *feedbackQuery) {
		q.osVersions = normalizeList(versions)
	}
}

// WithFeedbackAppPlatforms filters feedback by app platform(s).
func WithFeedbackAppPlatforms(platforms []string) FeedbackOption {
	return func(q *feedbackQuery) {
		q.appPlatforms = normalizeUpperList(platforms)
	}
}

// WithFeedbackDevicePlatforms filters feedback by device platform(s).
func WithFeedbackDevicePlatforms(platforms []string) FeedbackOption {
	return func(q *feedbackQuery) {
		q.devicePlatforms = normalizeUpperList(platforms)
	}
}

// WithFeedbackBuildIDs filters feedback by build ID(s).
func WithFeedbackBuildIDs(ids []string) FeedbackOption {
	return func(q *feedbackQuery) {
		q.buildIDs = normalizeList(ids)
	}
}

// WithFeedbackBuildPreReleaseVersionIDs filters feedback by pre-release version ID(s).
func WithFeedbackBuildPreReleaseVersionIDs(ids []string) FeedbackOption {
	return func(q *feedbackQuery) {
		q.buildPreReleaseVersionIDs = normalizeList(ids)
	}
}

// WithFeedbackTesterIDs filters feedback by tester ID(s).
func WithFeedbackTesterIDs(ids []string) FeedbackOption {
	return func(q *feedbackQuery) {
		q.testerIDs = normalizeList(ids)
	}
}

// WithFeedbackLimit sets the max number of feedback items to return.
func WithFeedbackLimit(limit int) FeedbackOption {
	return func(q *feedbackQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithFeedbackNextURL uses a next page URL directly.
func WithFeedbackNextURL(next string) FeedbackOption {
	return func(q *feedbackQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithFeedbackSort sets the sort order for feedback.
func WithFeedbackSort(sort string) FeedbackOption {
	return func(q *feedbackQuery) {
		if strings.TrimSpace(sort) != "" {
			q.sort = strings.TrimSpace(sort)
		}
	}
}

// WithFeedbackIncludeScreenshots includes screenshot URLs in feedback responses.
func WithFeedbackIncludeScreenshots() FeedbackOption {
	return func(q *feedbackQuery) {
		q.includeScreenshots = true
	}
}

// WithCrashDeviceModels filters crashes by device model(s).
func WithCrashDeviceModels(models []string) CrashOption {
	return func(q *crashQuery) {
		q.deviceModels = normalizeList(models)
	}
}

// WithCrashOSVersions filters crashes by OS version(s).
func WithCrashOSVersions(versions []string) CrashOption {
	return func(q *crashQuery) {
		q.osVersions = normalizeList(versions)
	}
}

// WithCrashAppPlatforms filters crashes by app platform(s).
func WithCrashAppPlatforms(platforms []string) CrashOption {
	return func(q *crashQuery) {
		q.appPlatforms = normalizeUpperList(platforms)
	}
}

// WithCrashDevicePlatforms filters crashes by device platform(s).
func WithCrashDevicePlatforms(platforms []string) CrashOption {
	return func(q *crashQuery) {
		q.devicePlatforms = normalizeUpperList(platforms)
	}
}

// WithCrashBuildIDs filters crashes by build ID(s).
func WithCrashBuildIDs(ids []string) CrashOption {
	return func(q *crashQuery) {
		q.buildIDs = normalizeList(ids)
	}
}

// WithCrashBuildPreReleaseVersionIDs filters crashes by pre-release version ID(s).
func WithCrashBuildPreReleaseVersionIDs(ids []string) CrashOption {
	return func(q *crashQuery) {
		q.buildPreReleaseVersionIDs = normalizeList(ids)
	}
}

// WithSubscriptionOfferCodeOneTimeUseCodesLimit sets the max number of offer code batches to return.
func WithSubscriptionOfferCodeOneTimeUseCodesLimit(limit int) SubscriptionOfferCodeOneTimeUseCodesOption {
	return func(q *subscriptionOfferCodeOneTimeUseCodesQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithSubscriptionOfferCodeOneTimeUseCodesNextURL uses a next page URL directly.
func WithSubscriptionOfferCodeOneTimeUseCodesNextURL(next string) SubscriptionOfferCodeOneTimeUseCodesOption {
	return func(q *subscriptionOfferCodeOneTimeUseCodesQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithCrashTesterIDs filters crashes by tester ID(s).
func WithCrashTesterIDs(ids []string) CrashOption {
	return func(q *crashQuery) {
		q.testerIDs = normalizeList(ids)
	}
}

// WithCrashLimit sets the max number of crash items to return.
func WithCrashLimit(limit int) CrashOption {
	return func(q *crashQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithAccessibilityDeclarationsDeviceFamilies filters declarations by device family.
func WithAccessibilityDeclarationsDeviceFamilies(families []string) AccessibilityDeclarationsOption {
	return func(q *accessibilityDeclarationsQuery) {
		q.deviceFamilies = normalizeUpperList(families)
	}
}

// WithAccessibilityDeclarationsStates filters declarations by state.
func WithAccessibilityDeclarationsStates(states []string) AccessibilityDeclarationsOption {
	return func(q *accessibilityDeclarationsQuery) {
		q.states = normalizeUpperList(states)
	}
}

// WithAccessibilityDeclarationsFields includes specific fields.
func WithAccessibilityDeclarationsFields(fields []string) AccessibilityDeclarationsOption {
	return func(q *accessibilityDeclarationsQuery) {
		q.fields = normalizeList(fields)
	}
}

// WithAccessibilityDeclarationsLimit sets the max number of declarations to return.
func WithAccessibilityDeclarationsLimit(limit int) AccessibilityDeclarationsOption {
	return func(q *accessibilityDeclarationsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithAccessibilityDeclarationsNextURL uses a next page URL directly.
func WithAccessibilityDeclarationsNextURL(next string) AccessibilityDeclarationsOption {
	return func(q *accessibilityDeclarationsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithAppStoreReviewAttachmentsFields includes specific attachment fields.
func WithAppStoreReviewAttachmentsFields(fields []string) AppStoreReviewAttachmentsOption {
	return func(q *appStoreReviewAttachmentsQuery) {
		q.fieldsAttachments = normalizeList(fields)
	}
}

// WithAppStoreReviewAttachmentReviewDetailFields includes fields for review detail when included.
func WithAppStoreReviewAttachmentReviewDetailFields(fields []string) AppStoreReviewAttachmentsOption {
	return func(q *appStoreReviewAttachmentsQuery) {
		q.fieldsReviewDetails = normalizeList(fields)
	}
}

// WithAppStoreReviewAttachmentsInclude includes related resources.
func WithAppStoreReviewAttachmentsInclude(include []string) AppStoreReviewAttachmentsOption {
	return func(q *appStoreReviewAttachmentsQuery) {
		q.include = normalizeList(include)
	}
}

// WithAppStoreReviewAttachmentsLimit sets the max number of attachments to return.
func WithAppStoreReviewAttachmentsLimit(limit int) AppStoreReviewAttachmentsOption {
	return func(q *appStoreReviewAttachmentsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithAppStoreReviewAttachmentsNextURL uses a next page URL directly.
func WithAppStoreReviewAttachmentsNextURL(next string) AppStoreReviewAttachmentsOption {
	return func(q *appStoreReviewAttachmentsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithAppEncryptionDeclarationsBuildIDs filters declarations by build IDs.
func WithAppEncryptionDeclarationsBuildIDs(ids []string) AppEncryptionDeclarationsOption {
	return func(q *appEncryptionDeclarationsQuery) {
		q.buildIDs = normalizeList(ids)
	}
}

// WithAppEncryptionDeclarationsFields includes specific declaration fields.
func WithAppEncryptionDeclarationsFields(fields []string) AppEncryptionDeclarationsOption {
	return func(q *appEncryptionDeclarationsQuery) {
		q.fields = normalizeList(fields)
	}
}

// WithAppEncryptionDeclarationsDocumentFields includes document fields when included.
func WithAppEncryptionDeclarationsDocumentFields(fields []string) AppEncryptionDeclarationsOption {
	return func(q *appEncryptionDeclarationsQuery) {
		q.documentFields = normalizeList(fields)
	}
}

// WithAppEncryptionDeclarationsInclude includes related resources.
func WithAppEncryptionDeclarationsInclude(include []string) AppEncryptionDeclarationsOption {
	return func(q *appEncryptionDeclarationsQuery) {
		q.include = normalizeList(include)
	}
}

// WithAppEncryptionDeclarationsLimit sets the max number of declarations to return.
func WithAppEncryptionDeclarationsLimit(limit int) AppEncryptionDeclarationsOption {
	return func(q *appEncryptionDeclarationsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithAppEncryptionDeclarationsBuildLimit sets the max number of related builds when included.
func WithAppEncryptionDeclarationsBuildLimit(limit int) AppEncryptionDeclarationsOption {
	return func(q *appEncryptionDeclarationsQuery) {
		if limit > 0 {
			q.buildLimit = limit
		}
	}
}

// WithAppEncryptionDeclarationsNextURL uses a next page URL directly.
func WithAppEncryptionDeclarationsNextURL(next string) AppEncryptionDeclarationsOption {
	return func(q *appEncryptionDeclarationsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithCrashNextURL uses a next page URL directly.
func WithCrashNextURL(next string) CrashOption {
	return func(q *crashQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithCrashSort sets the sort order for crashes.
func WithCrashSort(sort string) CrashOption {
	return func(q *crashQuery) {
		if strings.TrimSpace(sort) != "" {
			q.sort = strings.TrimSpace(sort)
		}
	}
}

// WithRating filters reviews by star rating (1-5).
func WithRating(rating int) ReviewOption {
	return func(r *reviewQuery) {
		if rating >= 1 && rating <= 5 {
			r.rating = rating
		}
	}
}

// WithTerritory filters reviews by territory code (e.g. US, GBR).
func WithTerritory(territory string) ReviewOption {
	return func(r *reviewQuery) {
		if territory != "" {
			r.territory = strings.ToUpper(territory)
		}
	}
}

// WithReviewSort sets the sort order for reviews.
func WithReviewSort(sort string) ReviewOption {
	return func(r *reviewQuery) {
		if strings.TrimSpace(sort) != "" {
			r.sort = strings.TrimSpace(sort)
		}
	}
}

// WithLimit sets the max number of reviews to return.
func WithLimit(limit int) ReviewOption {
	return func(r *reviewQuery) {
		if limit > 0 {
			r.limit = limit
		}
	}
}

// WithNextURL uses a next page URL directly.
func WithNextURL(next string) ReviewOption {
	return func(r *reviewQuery) {
		if strings.TrimSpace(next) != "" {
			r.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithAppsLimit sets the max number of apps to return.
func WithAppsLimit(limit int) AppsOption {
	return func(q *appsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithAppsNextURL uses a next page URL directly.
func WithAppsNextURL(next string) AppsOption {
	return func(q *appsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithAppsSort sets the sort order for apps.
func WithAppsSort(sort string) AppsOption {
	return func(q *appsQuery) {
		if strings.TrimSpace(sort) != "" {
			q.sort = strings.TrimSpace(sort)
		}
	}
}

// WithAppsBundleIDs filters apps by bundle ID(s).
func WithAppsBundleIDs(bundleIDs []string) AppsOption {
	return func(q *appsQuery) {
		q.bundleIDs = normalizeList(bundleIDs)
	}
}

// WithAppsNames filters apps by name(s).
func WithAppsNames(names []string) AppsOption {
	return func(q *appsQuery) {
		q.names = normalizeList(names)
	}
}

// WithAppsSKUs filters apps by SKU(s).
func WithAppsSKUs(skus []string) AppsOption {
	return func(q *appsQuery) {
		q.skus = normalizeList(skus)
	}
}

// WithAppTagsLimit sets the max number of app tags to return.
func WithAppTagsLimit(limit int) AppTagsOption {
	return func(q *appTagsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithAppTagsNextURL uses a next page URL directly.
func WithAppTagsNextURL(next string) AppTagsOption {
	return func(q *appTagsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithAppTagsVisibleInAppStore filters app tags by visibility.
func WithAppTagsVisibleInAppStore(values []string) AppTagsOption {
	return func(q *appTagsQuery) {
		q.visibleInAppStore = normalizeList(values)
	}
}

// WithAppTagsSort sets the sort order for app tags.
func WithAppTagsSort(sort string) AppTagsOption {
	return func(q *appTagsQuery) {
		if strings.TrimSpace(sort) != "" {
			q.sort = strings.TrimSpace(sort)
		}
	}
}

// WithAppTagsFields sets fields[appTags] for app tag responses.
func WithAppTagsFields(fields []string) AppTagsOption {
	return func(q *appTagsQuery) {
		q.fields = normalizeList(fields)
	}
}

// WithAppTagsInclude sets include for app tag responses.
func WithAppTagsInclude(include []string) AppTagsOption {
	return func(q *appTagsQuery) {
		q.include = normalizeList(include)
	}
}

// WithAppTagsTerritoryFields sets fields[territories] for included territory responses.
func WithAppTagsTerritoryFields(fields []string) AppTagsOption {
	return func(q *appTagsQuery) {
		q.territoryFields = normalizeList(fields)
	}
}

// WithAppTagsTerritoryLimit sets limit[territories] for included territories.
func WithAppTagsTerritoryLimit(limit int) AppTagsOption {
	return func(q *appTagsQuery) {
		if limit > 0 {
			q.territoryLimit = limit
		}
	}
}

// WithBuildsLimit sets the max number of builds to return.
func WithBuildsLimit(limit int) BuildsOption {
	return func(q *buildsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBuildsNextURL uses a next page URL directly.
func WithBuildsNextURL(next string) BuildsOption {
	return func(q *buildsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithBuildsSort sets the sort order for builds.
func WithBuildsSort(sort string) BuildsOption {
	return func(q *buildsQuery) {
		if strings.TrimSpace(sort) != "" {
			q.sort = strings.TrimSpace(sort)
		}
	}
}

// WithBuildsPreReleaseVersion filters builds by pre-release version ID.
func WithBuildsPreReleaseVersion(preReleaseVersionID string) BuildsOption {
	return func(q *buildsQuery) {
		if strings.TrimSpace(preReleaseVersionID) != "" {
			q.preReleaseVersionID = strings.TrimSpace(preReleaseVersionID)
		}
	}
}

// WithBuildBundlesLimit sets the max number of included build bundles to return.
func WithBuildBundlesLimit(limit int) BuildBundlesOption {
	return func(q *buildBundlesQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBuildBundleFileSizesLimit sets the max number of file size items to return.
func WithBuildBundleFileSizesLimit(limit int) BuildBundleFileSizesOption {
	return func(q *buildBundleFileSizesQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBuildBundleFileSizesNextURL uses a next page URL directly.
func WithBuildBundleFileSizesNextURL(next string) BuildBundleFileSizesOption {
	return func(q *buildBundleFileSizesQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithBetaAppClipInvocationsLimit sets the max number of App Clip invocations to return.
func WithBetaAppClipInvocationsLimit(limit int) BetaAppClipInvocationsOption {
	return func(q *betaAppClipInvocationsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBetaAppClipInvocationsNextURL uses a next page URL directly.
func WithBetaAppClipInvocationsNextURL(next string) BetaAppClipInvocationsOption {
	return func(q *betaAppClipInvocationsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithAppStoreVersionsLimit sets the max number of versions to return.
func WithAppStoreVersionsLimit(limit int) AppStoreVersionsOption {
	return func(q *appStoreVersionsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithAppStoreVersionsNextURL uses a next page URL directly.
func WithAppStoreVersionsNextURL(next string) AppStoreVersionsOption {
	return func(q *appStoreVersionsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithAppStoreVersionsPlatforms filters versions by platform.
func WithAppStoreVersionsPlatforms(platforms []string) AppStoreVersionsOption {
	return func(q *appStoreVersionsQuery) {
		q.platforms = normalizeUpperList(platforms)
	}
}

// WithAppStoreVersionsVersionStrings filters versions by version string.
func WithAppStoreVersionsVersionStrings(versions []string) AppStoreVersionsOption {
	return func(q *appStoreVersionsQuery) {
		q.versionStrings = normalizeList(versions)
	}
}

// WithAppStoreVersionsStates filters versions by app store state.
func WithAppStoreVersionsStates(states []string) AppStoreVersionsOption {
	return func(q *appStoreVersionsQuery) {
		q.states = normalizeUpperList(states)
	}
}

// WithReviewSubmissionsLimit sets the max number of review submissions to return.
func WithReviewSubmissionsLimit(limit int) ReviewSubmissionsOption {
	return func(q *reviewSubmissionsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithReviewSubmissionsNextURL uses a next page URL directly.
func WithReviewSubmissionsNextURL(next string) ReviewSubmissionsOption {
	return func(q *reviewSubmissionsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithReviewSubmissionsPlatforms filters review submissions by platform.
func WithReviewSubmissionsPlatforms(platforms []string) ReviewSubmissionsOption {
	return func(q *reviewSubmissionsQuery) {
		q.platforms = normalizeUpperList(platforms)
	}
}

// WithReviewSubmissionsStates filters review submissions by state.
func WithReviewSubmissionsStates(states []string) ReviewSubmissionsOption {
	return func(q *reviewSubmissionsQuery) {
		q.states = normalizeUpperList(states)
	}
}

// WithReviewSubmissionItemsLimit sets the max number of review submission items to return.
func WithReviewSubmissionItemsLimit(limit int) ReviewSubmissionItemsOption {
	return func(q *reviewSubmissionItemsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithReviewSubmissionItemsNextURL uses a next page URL directly.
func WithReviewSubmissionItemsNextURL(next string) ReviewSubmissionItemsOption {
	return func(q *reviewSubmissionItemsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithPreReleaseVersionsPlatform filters pre-release versions by platform.
func WithPreReleaseVersionsPlatform(platform string) PreReleaseVersionsOption {
	return func(q *preReleaseVersionsQuery) {
		normalized := normalizeUpperCSVString(platform)
		if normalized != "" {
			q.platform = normalized
		}
	}
}

// WithPreReleaseVersionsVersion filters pre-release versions by version string.
func WithPreReleaseVersionsVersion(version string) PreReleaseVersionsOption {
	return func(q *preReleaseVersionsQuery) {
		normalized := normalizeCSVString(version)
		if normalized != "" {
			q.version = normalized
		}
	}
}

// WithPreReleaseVersionsLimit sets the max number of pre-release versions to return.
func WithPreReleaseVersionsLimit(limit int) PreReleaseVersionsOption {
	return func(q *preReleaseVersionsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithPreReleaseVersionsNextURL uses a next page URL directly.
func WithPreReleaseVersionsNextURL(next string) PreReleaseVersionsOption {
	return func(q *preReleaseVersionsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithBetaGroupsLimit sets the max number of beta groups to return.
func WithBetaGroupsLimit(limit int) BetaGroupsOption {
	return func(q *betaGroupsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBetaGroupsNextURL uses a next page URL directly.
func WithBetaGroupsNextURL(next string) BetaGroupsOption {
	return func(q *betaGroupsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithBetaGroupBuildsLimit sets the max number of builds to return for a group.
func WithBetaGroupBuildsLimit(limit int) BetaGroupBuildsOption {
	return func(q *betaGroupBuildsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBetaGroupBuildsNextURL uses a next page URL directly.
func WithBetaGroupBuildsNextURL(next string) BetaGroupBuildsOption {
	return func(q *betaGroupBuildsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithBetaGroupTestersLimit sets the max number of testers to return for a group.
func WithBetaGroupTestersLimit(limit int) BetaGroupTestersOption {
	return func(q *betaGroupTestersQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBetaGroupTestersNextURL uses a next page URL directly.
func WithBetaGroupTestersNextURL(next string) BetaGroupTestersOption {
	return func(q *betaGroupTestersQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithBetaTestersLimit sets the max number of beta testers to return.
func WithBetaTestersLimit(limit int) BetaTestersOption {
	return func(q *betaTestersQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBetaTestersNextURL uses a next page URL directly.
func WithBetaTestersNextURL(next string) BetaTestersOption {
	return func(q *betaTestersQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithBetaTestersEmail filters beta testers by email.
func WithBetaTestersEmail(email string) BetaTestersOption {
	return func(q *betaTestersQuery) {
		q.email = strings.TrimSpace(email)
	}
}

// WithBetaTestersGroupIDs filters beta testers by beta group ID(s).
func WithBetaTestersGroupIDs(ids []string) BetaTestersOption {
	return func(q *betaTestersQuery) {
		q.groupIDs = normalizeList(ids)
	}
}

// WithBetaTestersBuildID filters beta testers by build ID.
func WithBetaTestersBuildID(buildID string) BetaTestersOption {
	return func(q *betaTestersQuery) {
		q.filterBuilds = strings.TrimSpace(buildID)
	}
}

// WithBundleIDsLimit sets the max number of bundle IDs to return.
func WithBundleIDsLimit(limit int) BundleIDsOption {
	return func(q *bundleIDsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBundleIDsNextURL uses a next page URL directly.
func WithBundleIDsNextURL(next string) BundleIDsOption {
	return func(q *bundleIDsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithBundleIDsFilterIdentifier filters bundle IDs by identifier (supports CSV).
func WithBundleIDsFilterIdentifier(identifier string) BundleIDsOption {
	return func(q *bundleIDsQuery) {
		normalized := normalizeCSVString(identifier)
		if normalized != "" {
			q.identifier = normalized
		}
	}
}

// WithBundleIDCapabilitiesLimit sets the max number of capabilities to return.
func WithBundleIDCapabilitiesLimit(limit int) BundleIDCapabilitiesOption {
	return func(q *bundleIDCapabilitiesQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBundleIDCapabilitiesNextURL uses a next page URL directly.
func WithBundleIDCapabilitiesNextURL(next string) BundleIDCapabilitiesOption {
	return func(q *bundleIDCapabilitiesQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithCertificatesLimit sets the max number of certificates to return.
func WithCertificatesLimit(limit int) CertificatesOption {
	return func(q *certificatesQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithCertificatesNextURL uses a next page URL directly.
func WithCertificatesNextURL(next string) CertificatesOption {
	return func(q *certificatesQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithCertificatesTypes filters certificates by type.
func WithCertificatesTypes(types []string) CertificatesOption {
	return func(q *certificatesQuery) {
		q.certificateTypes = normalizeUpperList(types)
	}
}

// WithCertificatesFilterType filters certificates by certificate type (supports CSV).
func WithCertificatesFilterType(certType string) CertificatesOption {
	return func(q *certificatesQuery) {
		normalized := normalizeUpperCSVString(certType)
		if normalized == "" {
			return
		}
		q.certificateTypes = strings.Split(normalized, ",")
	}
}

// WithProfilesLimit sets the max number of profiles to return.
func WithProfilesLimit(limit int) ProfilesOption {
	return func(q *profilesQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithProfilesNextURL uses a next page URL directly.
func WithProfilesNextURL(next string) ProfilesOption {
	return func(q *profilesQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithProfilesTypes filters profiles by profile type.
func WithProfilesTypes(types []string) ProfilesOption {
	return func(q *profilesQuery) {
		q.profileTypes = normalizeUpperList(types)
	}
}

// WithProfilesFilterBundleID filters profiles by bundle ID.
func WithProfilesFilterBundleID(bundleID string) ProfilesOption {
	return func(q *profilesQuery) {
		if strings.TrimSpace(bundleID) != "" {
			q.bundleID = strings.TrimSpace(bundleID)
		}
	}
}

// WithProfilesFilterType filters profiles by profile type (supports CSV).
func WithProfilesFilterType(profileType string) ProfilesOption {
	return func(q *profilesQuery) {
		normalized := normalizeUpperCSVString(profileType)
		if normalized == "" {
			return
		}
		q.profileTypes = strings.Split(normalized, ",")
	}
}

// WithUsersLimit sets the max number of users to return.
func WithUsersLimit(limit int) UsersOption {
	return func(q *usersQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithUsersNextURL uses a next page URL directly.
func WithUsersNextURL(next string) UsersOption {
	return func(q *usersQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithUsersEmail filters users by email/username.
func WithUsersEmail(email string) UsersOption {
	return func(q *usersQuery) {
		q.email = strings.TrimSpace(email)
	}
}

// WithUsersRoles filters users by roles.
func WithUsersRoles(roles []string) UsersOption {
	return func(q *usersQuery) {
		q.roles = normalizeList(roles)
	}
}

// WithDevicesLimit sets the max number of devices to return.
func WithDevicesLimit(limit int) DevicesOption {
	return func(q *devicesQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithDevicesFilterUDIDs filters devices by UDID(s).
func WithDevicesFilterUDIDs(udids []string) DevicesOption {
	return WithDevicesUDIDs(udids)
}

// WithDevicesFilterPlatforms filters devices by platform(s).
func WithDevicesFilterPlatforms(platforms []string) DevicesOption {
	return func(q *devicesQuery) {
		normalized := normalizeUpperList(platforms)
		if len(normalized) == 0 {
			return
		}
		q.platforms = normalized
	}
}

// WithDevicesFilterStatuses filters devices by status (e.g., ENABLED, DISABLED).
func WithDevicesFilterStatuses(statuses []string) DevicesOption {
	return func(q *devicesQuery) {
		normalized := normalizeUpperList(statuses)
		if len(normalized) == 0 {
			return
		}
		q.status = strings.Join(normalized, ",")
	}
}

// WithDevicesNextURL uses a next page URL directly.
func WithDevicesNextURL(next string) DevicesOption {
	return func(q *devicesQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithDevicesNames filters devices by name(s).
func WithDevicesNames(names []string) DevicesOption {
	return func(q *devicesQuery) {
		q.names = normalizeList(names)
	}
}

// WithDevicesPlatform filters devices by platform.
func WithDevicesPlatform(platform string) DevicesOption {
	return func(q *devicesQuery) {
		if strings.TrimSpace(platform) != "" {
			q.platforms = []string{strings.ToUpper(strings.TrimSpace(platform))}
		}
	}
}

// WithDevicesPlatforms filters devices by platform(s).
func WithDevicesPlatforms(platforms []string) DevicesOption {
	return func(q *devicesQuery) {
		q.platforms = normalizeUpperList(platforms)
	}
}

// WithDevicesStatus filters devices by status.
func WithDevicesStatus(status string) DevicesOption {
	return func(q *devicesQuery) {
		if strings.TrimSpace(status) != "" {
			q.status = strings.ToUpper(strings.TrimSpace(status))
		}
	}
}

// WithDevicesUDIDs filters devices by UDID(s).
func WithDevicesUDIDs(udids []string) DevicesOption {
	return func(q *devicesQuery) {
		q.udids = normalizeList(udids)
	}
}

// WithDevicesIDs filters devices by ID(s).
func WithDevicesIDs(ids []string) DevicesOption {
	return func(q *devicesQuery) {
		q.ids = normalizeList(ids)
	}
}

// WithDevicesSort sets the sort order for devices.
func WithDevicesSort(sort string) DevicesOption {
	return func(q *devicesQuery) {
		if strings.TrimSpace(sort) != "" {
			q.sort = strings.TrimSpace(sort)
		}
	}
}

// WithDevicesFields sets fields[devices] for device responses.
func WithDevicesFields(fields []string) DevicesOption {
	return func(q *devicesQuery) {
		q.fields = normalizeList(fields)
	}
}

// WithUserInvitationsLimit sets the max number of invitations to return.
func WithUserInvitationsLimit(limit int) UserInvitationsOption {
	return func(q *userInvitationsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithUserInvitationsNextURL uses a next page URL directly.
func WithUserInvitationsNextURL(next string) UserInvitationsOption {
	return func(q *userInvitationsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithBetaAppReviewDetailsLimit sets the max number of review detail records to return.
func WithBetaAppReviewDetailsLimit(limit int) BetaAppReviewDetailsOption {
	return func(q *betaAppReviewDetailsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBetaAppReviewDetailsNextURL uses a next page URL directly.
func WithBetaAppReviewDetailsNextURL(next string) BetaAppReviewDetailsOption {
	return func(q *betaAppReviewDetailsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithBetaAppReviewSubmissionsLimit sets the max number of submissions to return.
func WithBetaAppReviewSubmissionsLimit(limit int) BetaAppReviewSubmissionsOption {
	return func(q *betaAppReviewSubmissionsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBetaAppReviewSubmissionsNextURL uses a next page URL directly.
func WithBetaAppReviewSubmissionsNextURL(next string) BetaAppReviewSubmissionsOption {
	return func(q *betaAppReviewSubmissionsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithBetaAppReviewSubmissionsBuildIDs filters submissions by build ID(s).
func WithBetaAppReviewSubmissionsBuildIDs(ids []string) BetaAppReviewSubmissionsOption {
	return func(q *betaAppReviewSubmissionsQuery) {
		q.buildIDs = normalizeList(ids)
	}
}

// WithBuildBetaDetailsLimit sets the max number of build beta details to return.
func WithBuildBetaDetailsLimit(limit int) BuildBetaDetailsOption {
	return func(q *buildBetaDetailsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBuildBetaDetailsNextURL uses a next page URL directly.
func WithBuildBetaDetailsNextURL(next string) BuildBetaDetailsOption {
	return func(q *buildBetaDetailsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithBuildBetaDetailsBuildIDs filters build beta details by build ID(s).
func WithBuildBetaDetailsBuildIDs(ids []string) BuildBetaDetailsOption {
	return func(q *buildBetaDetailsQuery) {
		q.buildIDs = normalizeList(ids)
	}
}

// WithBetaRecruitmentCriterionOptionsLimit sets the max number of criterion options to return.
func WithBetaRecruitmentCriterionOptionsLimit(limit int) BetaRecruitmentCriterionOptionsOption {
	return func(q *betaRecruitmentCriterionOptionsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBetaRecruitmentCriterionOptionsNextURL uses a next page URL directly.
func WithBetaRecruitmentCriterionOptionsNextURL(next string) BetaRecruitmentCriterionOptionsOption {
	return func(q *betaRecruitmentCriterionOptionsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithAppStoreVersionLocalizationsLimit sets the max number of localizations to return.
func WithAppStoreVersionLocalizationsLimit(limit int) AppStoreVersionLocalizationsOption {
	return func(q *appStoreVersionLocalizationsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithAppStoreVersionLocalizationsNextURL uses a next page URL directly.
func WithAppStoreVersionLocalizationsNextURL(next string) AppStoreVersionLocalizationsOption {
	return func(q *appStoreVersionLocalizationsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithAppStoreVersionLocalizationLocales filters version localizations by locale.
func WithAppStoreVersionLocalizationLocales(locales []string) AppStoreVersionLocalizationsOption {
	return func(q *appStoreVersionLocalizationsQuery) {
		q.locales = normalizeList(locales)
	}
}

// WithBetaBuildLocalizationsLimit sets the max number of beta build localizations to return.
func WithBetaBuildLocalizationsLimit(limit int) BetaBuildLocalizationsOption {
	return func(q *betaBuildLocalizationsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBetaBuildLocalizationsNextURL uses a next page URL directly.
func WithBetaBuildLocalizationsNextURL(next string) BetaBuildLocalizationsOption {
	return func(q *betaBuildLocalizationsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithBetaBuildLocalizationLocales filters beta build localizations by locale.
func WithBetaBuildLocalizationLocales(locales []string) BetaBuildLocalizationsOption {
	return func(q *betaBuildLocalizationsQuery) {
		q.locales = normalizeList(locales)
	}
}

// WithAppInfoLocalizationsLimit sets the max number of app info localizations to return.
func WithAppInfoLocalizationsLimit(limit int) AppInfoLocalizationsOption {
	return func(q *appInfoLocalizationsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithAppInfoLocalizationsNextURL uses a next page URL directly.
func WithAppInfoLocalizationsNextURL(next string) AppInfoLocalizationsOption {
	return func(q *appInfoLocalizationsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithAppInfoLocalizationLocales filters app info localizations by locale.
func WithAppInfoLocalizationLocales(locales []string) AppInfoLocalizationsOption {
	return func(q *appInfoLocalizationsQuery) {
		q.locales = normalizeList(locales)
	}
}

// WithTerritoriesLimit sets the max number of territories to return.
func WithTerritoriesLimit(limit int) TerritoriesOption {
	return func(q *territoriesQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithTerritoriesNextURL uses a next page URL directly.
func WithTerritoriesNextURL(next string) TerritoriesOption {
	return func(q *territoriesQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithTerritoriesFields sets fields[territories] for territory responses.
func WithTerritoriesFields(fields []string) TerritoriesOption {
	return func(q *territoriesQuery) {
		q.fields = normalizeList(fields)
	}
}

// WithLinkagesLimit sets the max number of linkages to return.
func WithLinkagesLimit(limit int) LinkagesOption {
	return func(q *linkagesQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithLinkagesNextURL uses a next page URL directly.
func WithLinkagesNextURL(next string) LinkagesOption {
	return func(q *linkagesQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithPricePointsLimit sets the max number of price points to return.
func WithPricePointsLimit(limit int) PricePointsOption {
	return func(q *pricePointsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithPricePointsNextURL uses a next page URL directly.
func WithPricePointsNextURL(next string) PricePointsOption {
	return func(q *pricePointsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithPricePointsTerritory filters app price points by territory.
func WithPricePointsTerritory(territory string) PricePointsOption {
	return func(q *pricePointsQuery) {
		if strings.TrimSpace(territory) != "" {
			q.territory = strings.ToUpper(strings.TrimSpace(territory))
		}
	}
}
