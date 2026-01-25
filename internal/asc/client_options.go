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

// BuildsOption is a functional option for GetBuilds.
type BuildsOption func(*buildsQuery)

// AppStoreVersionsOption is a functional option for GetAppStoreVersions.
type AppStoreVersionsOption func(*appStoreVersionsQuery)

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

// UsersOption is a functional option for GetUsers.
type UsersOption func(*usersQuery)

// UserInvitationsOption is a functional option for GetUserInvitations.
type UserInvitationsOption func(*userInvitationsQuery)

// AppStoreVersionLocalizationsOption is a functional option for version localizations.
type AppStoreVersionLocalizationsOption func(*appStoreVersionLocalizationsQuery)

// AppInfoLocalizationsOption is a functional option for app info localizations.
type AppInfoLocalizationsOption func(*appInfoLocalizationsQuery)

// TerritoriesOption is a functional option for GetTerritories.
type TerritoriesOption func(*territoriesQuery)

// PricePointsOption is a functional option for GetAppPricePoints.
type PricePointsOption func(*pricePointsQuery)

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
