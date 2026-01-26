package asc

import (
	"fmt"
	"net/url"
	"strings"
)

type listQuery struct {
	limit   int
	nextURL string
}

type feedbackQuery struct {
	listQuery
	deviceModels              []string
	osVersions                []string
	appPlatforms              []string
	devicePlatforms           []string
	buildIDs                  []string
	buildPreReleaseVersionIDs []string
	testerIDs                 []string
	sort                      string
	includeScreenshots        bool
}

type crashQuery struct {
	listQuery
	deviceModels              []string
	osVersions                []string
	appPlatforms              []string
	devicePlatforms           []string
	buildIDs                  []string
	buildPreReleaseVersionIDs []string
	testerIDs                 []string
	sort                      string
}

type reviewQuery struct {
	listQuery
	rating    int
	territory string
	sort      string
}

type appsQuery struct {
	listQuery
	sort      string
	bundleIDs []string
	names     []string
	skus      []string
}

type buildsQuery struct {
	listQuery
	sort                string
	preReleaseVersionID string
}

type subscriptionOfferCodeOneTimeUseCodesQuery struct {
	listQuery
}

type appStoreVersionsQuery struct {
	listQuery
	platforms      []string
	versionStrings []string
	states         []string
}

type preReleaseVersionsQuery struct {
	listQuery
	platform string
	version  string
}

type appStoreVersionLocalizationsQuery struct {
	listQuery
	locales []string
}

type betaBuildLocalizationsQuery struct {
	listQuery
	locales []string
}

type appInfoLocalizationsQuery struct {
	listQuery
	locales []string
}

type betaGroupsQuery struct {
	listQuery
}

type betaGroupBuildsQuery struct {
	listQuery
}

type betaGroupTestersQuery struct {
	listQuery
}

type betaTestersQuery struct {
	listQuery
	email        string
	groupIDs     []string
	filterBuilds string
}

type bundleIDsQuery struct {
	listQuery
}

type bundleIDCapabilitiesQuery struct {
	listQuery
}

type certificatesQuery struct {
	listQuery
	certificateTypes []string
}

type profilesQuery struct {
	listQuery
	profileTypes []string
}

type usersQuery struct {
	listQuery
	email string
	roles []string
}

type devicesQuery struct {
	listQuery
	names     []string
	platforms []string
	status    string
	udids     []string
	ids       []string
	sort      string
	fields    []string
}

type userInvitationsQuery struct {
	listQuery
}

type territoriesQuery struct {
	listQuery
}

type pricePointsQuery struct {
	listQuery
	territory string
}

type betaAppReviewDetailsQuery struct {
	listQuery
}

type betaAppReviewSubmissionsQuery struct {
	listQuery
	buildIDs []string
}

type buildBetaDetailsQuery struct {
	listQuery
	buildIDs []string
}

type betaRecruitmentCriterionOptionsQuery struct {
	listQuery
}

func buildReviewQuery(opts []ReviewOption) string {
	query := &reviewQuery{}
	for _, opt := range opts {
		opt(query)
	}

	values := url.Values{}
	if query.territory != "" {
		values.Set("filter[territory]", query.territory)
	}
	if query.rating >= 1 && query.rating <= 5 {
		values.Set("filter[rating]", fmt.Sprintf("%d", query.rating))
	}
	if query.sort != "" {
		values.Set("sort", query.sort)
	}
	addLimit(values, query.limit)

	return values.Encode()
}

func buildAppsQuery(query *appsQuery) string {
	values := url.Values{}
	addCSV(values, "filter[bundleId]", query.bundleIDs)
	addCSV(values, "filter[name]", query.names)
	addCSV(values, "filter[sku]", query.skus)
	if query.sort != "" {
		values.Set("sort", query.sort)
	}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildFeedbackQuery(query *feedbackQuery) string {
	values := url.Values{}
	if query.includeScreenshots {
		values.Set("fields[betaFeedbackScreenshotSubmissions]", strings.Join([]string{
			"createdDate",
			"comment",
			"email",
			"deviceModel",
			"osVersion",
			"appPlatform",
			"devicePlatform",
			"screenshots",
		}, ","))
	}
	addCSV(values, "filter[deviceModel]", query.deviceModels)
	addCSV(values, "filter[osVersion]", query.osVersions)
	addCSV(values, "filter[appPlatform]", query.appPlatforms)
	addCSV(values, "filter[devicePlatform]", query.devicePlatforms)
	addCSV(values, "filter[build]", query.buildIDs)
	addCSV(values, "filter[build.preReleaseVersion]", query.buildPreReleaseVersionIDs)
	addCSV(values, "filter[tester]", query.testerIDs)
	if query.sort != "" {
		values.Set("sort", query.sort)
	}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildCrashQuery(query *crashQuery) string {
	values := url.Values{}
	addCSV(values, "filter[deviceModel]", query.deviceModels)
	addCSV(values, "filter[osVersion]", query.osVersions)
	addCSV(values, "filter[appPlatform]", query.appPlatforms)
	addCSV(values, "filter[devicePlatform]", query.devicePlatforms)
	addCSV(values, "filter[build]", query.buildIDs)
	addCSV(values, "filter[build.preReleaseVersion]", query.buildPreReleaseVersionIDs)
	addCSV(values, "filter[tester]", query.testerIDs)
	if query.sort != "" {
		values.Set("sort", query.sort)
	}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildBetaGroupsQuery(query *betaGroupsQuery) string {
	values := url.Values{}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildBetaGroupBuildsQuery(query *betaGroupBuildsQuery) string {
	values := url.Values{}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildBetaGroupTestersQuery(query *betaGroupTestersQuery) string {
	values := url.Values{}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildBetaTestersQuery(appID string, query *betaTestersQuery) string {
	values := url.Values{}
	// API allows only one relationship filter, so prefer builds over apps if provided
	if strings.TrimSpace(query.filterBuilds) != "" {
		values.Set("filter[builds]", strings.TrimSpace(query.filterBuilds))
	} else if strings.TrimSpace(appID) != "" {
		values.Set("filter[apps]", strings.TrimSpace(appID))
	}
	if strings.TrimSpace(query.email) != "" {
		values.Set("filter[email]", strings.TrimSpace(query.email))
	}
	addCSV(values, "filter[betaGroups]", query.groupIDs)
	addLimit(values, query.limit)
	return values.Encode()
}

func buildBundleIDsQuery(query *bundleIDsQuery) string {
	values := url.Values{}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildBundleIDCapabilitiesQuery(_ *bundleIDCapabilitiesQuery) string {
	// Bundle ID capabilities endpoint does not support limit/pagination params.
	return ""
}

func buildCertificatesQuery(query *certificatesQuery) string {
	values := url.Values{}
	addCSV(values, "filter[certificateType]", query.certificateTypes)
	addLimit(values, query.limit)
	return values.Encode()
}

func buildProfilesQuery(query *profilesQuery) string {
	values := url.Values{}
	addCSV(values, "filter[profileType]", query.profileTypes)
	addLimit(values, query.limit)
	return values.Encode()
}

func buildUsersQuery(query *usersQuery) string {
	values := url.Values{}
	if strings.TrimSpace(query.email) != "" {
		values.Set("filter[username]", strings.TrimSpace(query.email))
	}
	addCSV(values, "filter[roles]", query.roles)
	addLimit(values, query.limit)
	return values.Encode()
}

func buildDevicesQuery(query *devicesQuery) string {
	values := url.Values{}
	addCSV(values, "filter[name]", query.names)
	addCSV(values, "filter[platform]", query.platforms)
	if strings.TrimSpace(query.status) != "" {
		values.Set("filter[status]", strings.TrimSpace(query.status))
	}
	addCSV(values, "filter[udid]", query.udids)
	addCSV(values, "filter[id]", query.ids)
	if query.sort != "" {
		values.Set("sort", query.sort)
	}
	addCSV(values, "fields[devices]", query.fields)
	addLimit(values, query.limit)
	return values.Encode()
}

func buildDevicesFieldsQuery(fields []string) string {
	values := url.Values{}
	addCSV(values, "fields[devices]", fields)
	return values.Encode()
}

func buildUserInvitationsQuery(query *userInvitationsQuery) string {
	values := url.Values{}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildBetaAppReviewDetailsQuery(appID string, query *betaAppReviewDetailsQuery) string {
	values := url.Values{}
	if strings.TrimSpace(appID) != "" {
		values.Set("filter[app]", strings.TrimSpace(appID))
	}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildBetaAppReviewSubmissionsQuery(query *betaAppReviewSubmissionsQuery) string {
	values := url.Values{}
	addCSV(values, "filter[build]", query.buildIDs)
	addLimit(values, query.limit)
	return values.Encode()
}

func buildBuildBetaDetailsQuery(query *buildBetaDetailsQuery) string {
	values := url.Values{}
	addCSV(values, "filter[build]", query.buildIDs)
	addLimit(values, query.limit)
	return values.Encode()
}

func buildBetaRecruitmentCriterionOptionsQuery(query *betaRecruitmentCriterionOptionsQuery) string {
	values := url.Values{}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildSubscriptionOfferCodeOneTimeUseCodesQuery(query *subscriptionOfferCodeOneTimeUseCodesQuery) string {
	values := url.Values{}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildAppStoreVersionsQuery(query *appStoreVersionsQuery) string {
	values := url.Values{}
	addCSV(values, "filter[platform]", query.platforms)
	addCSV(values, "filter[versionString]", query.versionStrings)
	addCSV(values, "filter[appStoreState]", query.states)
	addLimit(values, query.limit)
	return values.Encode()
}

func buildPreReleaseVersionsQuery(appID string, query *preReleaseVersionsQuery) string {
	values := url.Values{}
	if strings.TrimSpace(appID) != "" {
		values.Set("filter[app]", strings.TrimSpace(appID))
	}
	if strings.TrimSpace(query.platform) != "" {
		values.Set("filter[platform]", strings.TrimSpace(query.platform))
	}
	if strings.TrimSpace(query.version) != "" {
		values.Set("filter[version]", strings.TrimSpace(query.version))
	}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildAppStoreVersionLocalizationsQuery(query *appStoreVersionLocalizationsQuery) string {
	values := url.Values{}
	addCSV(values, "filter[locale]", query.locales)
	addLimit(values, query.limit)
	return values.Encode()
}

func buildBetaBuildLocalizationsQuery(query *betaBuildLocalizationsQuery) string {
	values := url.Values{}
	addCSV(values, "filter[locale]", query.locales)
	addLimit(values, query.limit)
	return values.Encode()
}

func buildAppInfoLocalizationsQuery(query *appInfoLocalizationsQuery) string {
	values := url.Values{}
	addCSV(values, "filter[locale]", query.locales)
	addLimit(values, query.limit)
	return values.Encode()
}

func buildTerritoriesQuery(query *territoriesQuery) string {
	values := url.Values{}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildPricePointsQuery(query *pricePointsQuery) string {
	values := url.Values{}
	if strings.TrimSpace(query.territory) != "" {
		values.Set("filter[territory]", strings.TrimSpace(query.territory))
	}
	addLimit(values, query.limit)
	return values.Encode()
}
