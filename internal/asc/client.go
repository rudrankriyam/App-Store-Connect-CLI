package asc

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/auth"
)

const (
	// BaseURL is the App Store Connect API base URL
	BaseURL = "https://api.appstoreconnect.apple.com"
	// DefaultTimeout is the default request timeout
	DefaultTimeout = 30 * time.Second
	tokenLifetime  = 20 * time.Minute
)

// Client is an App Store Connect API client
type Client struct {
	httpClient *http.Client
	keyID      string
	issuerID   string
	privateKey *ecdsa.PrivateKey
}

// Resource is a generic ASC API resource wrapper.
type Resource[T any] struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	Attributes T      `json:"attributes"`
}

// Response is a generic ASC API response wrapper.
type Response[T any] struct {
	Data  []Resource[T] `json:"data"`
	Links Links         `json:"links,omitempty"`
}

// FeedbackAttributes describes beta feedback screenshot submissions.
type FeedbackAttributes struct {
	CreatedDate    string `json:"createdDate"`
	Comment        string `json:"comment"`
	Email          string `json:"email"`
	DeviceModel    string `json:"deviceModel,omitempty"`
	OSVersion      string `json:"osVersion,omitempty"`
	AppPlatform    string `json:"appPlatform,omitempty"`
	DevicePlatform string `json:"devicePlatform,omitempty"`
}

// CrashAttributes describes beta feedback crash submissions.
type CrashAttributes struct {
	CreatedDate    string `json:"createdDate"`
	Comment        string `json:"comment"`
	Email          string `json:"email"`
	DeviceModel    string `json:"deviceModel,omitempty"`
	OSVersion      string `json:"osVersion,omitempty"`
	AppPlatform    string `json:"appPlatform,omitempty"`
	DevicePlatform string `json:"devicePlatform,omitempty"`
	CrashLog       string `json:"crashLog,omitempty"`
}

// ReviewAttributes describes App Store customer reviews.
type ReviewAttributes struct {
	Rating           int    `json:"rating"`
	Title            string `json:"title"`
	Body             string `json:"body"`
	ReviewerNickname string `json:"reviewerNickname"`
	CreatedDate      string `json:"createdDate"`
	Territory        string `json:"territory"`
}

// FeedbackResponse is the response from beta feedback screenshots endpoint.
type FeedbackResponse = Response[FeedbackAttributes]

// CrashesResponse is the response from beta feedback crashes endpoint.
type CrashesResponse = Response[CrashAttributes]

// ReviewsResponse is the response from customer reviews endpoint.
type ReviewsResponse = Response[ReviewAttributes]

// AppsResponse is the response from apps endpoint.
type AppsResponse = Response[AppAttributes]

// BuildsResponse is the response from builds endpoint.
type BuildsResponse = Response[BuildAttributes]

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
	sort string
}

type buildsQuery struct {
	listQuery
	appID string
	sort string
}

// AppAttributes describes an app resource.
type AppAttributes struct {
	Name          string `json:"name"`
	BundleID      string `json:"bundleId"`
	SKU           string `json:"sku"`
	PrimaryLocale string `json:"primaryLocale,omitempty"`
}

// BuildAttributes describes a build resource.
type BuildAttributes struct {
	Version                 string `json:"version"`
	UploadedDate            string `json:"uploadedDate"`
	ExpirationDate          string `json:"expirationDate,omitempty"`
	ProcessingState         string `json:"processingState,omitempty"`
	MinOSVersion            string `json:"minOsVersion,omitempty"`
	UsesNonExemptEncryption bool   `json:"usesNonExemptEncryption,omitempty"`
	Expired                 bool   `json:"expired,omitempty"`
}

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

// WithBuildsLimit sets the max number of builds to return.
func WithBuildsLimit(limit int) BuildsOption {
	return func(q *buildsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBuildsApp filters builds by app ID.
func WithBuildsApp(appID string) BuildsOption {
	return func(q *buildsQuery) {
		q.appID = strings.TrimSpace(appID)
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

// NewClient creates a new ASC client
func NewClient(keyID, issuerID, privateKeyPath string) (*Client, error) {
	if err := auth.ValidateKeyFile(privateKeyPath); err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	key, err := auth.LoadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		keyID:      keyID,
		issuerID:   issuerID,
		privateKey: key,
	}, nil
}

// newRequest creates a new HTTP request with JWT authentication
func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	// Generate JWT token
	token, err := c.generateJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	url := path
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		url = BaseURL + path
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// generateJWT generates a JWT for ASC API authentication
func (c *Client) generateJWT() (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    c.issuerID,
		Audience:  jwt.ClaimStrings{"appstoreconnect-v1"},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenLifetime)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = c.keyID

	// Sign with the private key
	signedToken, err := token.SignedString(c.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// do performs an HTTP request and returns the response
func (c *Client) do(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		if err := ParseError(respBody); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
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

func buildFeedbackQuery(query *feedbackQuery) string {
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

func normalizeList(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		normalized = append(normalized, value)
	}
	return normalized
}

func normalizeUpperList(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		normalized = append(normalized, strings.ToUpper(value))
	}
	return normalized
}

func addCSV(values url.Values, key string, items []string) {
	items = normalizeList(items)
	if len(items) == 0 {
		return
	}
	values.Set(key, strings.Join(items, ","))
}

func addLimit(values url.Values, limit int) {
	if limit > 0 {
		values.Set("limit", strconv.Itoa(limit))
	}
}

// GetFeedback retrieves TestFlight feedback
func (c *Client) GetFeedback(ctx context.Context, appID string, opts ...FeedbackOption) (*FeedbackResponse, error) {
	query := &feedbackQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/apps/%s/betaFeedbackScreenshotSubmissions", appID)
	if query.nextURL != "" {
		path = query.nextURL
	} else if queryString := buildFeedbackQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response FeedbackResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetCrashes retrieves TestFlight crash reports
func (c *Client) GetCrashes(ctx context.Context, appID string, opts ...CrashOption) (*CrashesResponse, error) {
	query := &crashQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/apps/%s/betaFeedbackCrashSubmissions", appID)
	if query.nextURL != "" {
		path = query.nextURL
	} else if queryString := buildCrashQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response CrashesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetReviews retrieves App Store reviews
func (c *Client) GetReviews(ctx context.Context, appID string, opts ...ReviewOption) (*ReviewsResponse, error) {
	query := &reviewQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/apps/%s/customerReviews", appID)
	if query.nextURL != "" {
		path = query.nextURL
	} else if queryString := buildReviewQuery(opts); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response ReviewsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetApps retrieves the list of apps
func (c *Client) GetApps(ctx context.Context, opts ...AppsOption) (*AppsResponse, error) {
	query := &appsQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := "/v1/apps"
	if query.nextURL != "" {
		path = query.nextURL
	} else {
		values := url.Values{}
		if query.sort != "" {
			values.Set("sort", query.sort)
		}
		if query.limit > 0 {
			values.Set("limit", strconv.Itoa(query.limit))
		}
		if queryString := values.Encode(); queryString != "" {
			path += "?" + queryString
		}
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response AppsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetBuilds retrieves the list of builds for an app
func (c *Client) GetBuilds(ctx context.Context, appID string, opts ...BuildsOption) (*BuildsResponse, error) {
	query := &buildsQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/apps/%s/builds", appID)
	if query.nextURL != "" {
		path = query.nextURL
	} else {
		values := url.Values{}
		if query.sort != "" {
			values.Set("sort", query.sort)
		}
		if query.limit > 0 {
			values.Set("limit", strconv.Itoa(query.limit))
		}
		if queryString := values.Encode(); queryString != "" {
			path += "?" + queryString
		}
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response BuildsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// Links represents pagination links
type Links struct {
	Self string `json:"self,omitempty"`
	Next string `json:"next,omitempty"`
	Prev string `json:"prev,omitempty"`
}

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
	case *BuildsResponse:
		return printBuildsMarkdown(v)
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
	case *BuildsResponse:
		return printBuildsTable(v)
	default:
		return PrintJSON(data)
	}
}

// BuildRequestBody builds a JSON request body
func BuildRequestBody(data interface{}) (io.Reader, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}
	return &buf, nil
}

// ParseError parses an error response
func ParseError(body []byte) error {
	var errResp struct {
		Errors []struct {
			Code   string `json:"code"`
			Title  string `json:"title"`
			Detail string `json:"detail"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &errResp); err == nil && len(errResp.Errors) > 0 {
		return fmt.Errorf("%s: %s", errResp.Errors[0].Title, errResp.Errors[0].Detail)
	}

	return fmt.Errorf("unknown error: %s", string(body))
}

// IsNotFound checks if the error is a "not found" error
func IsNotFound(err error) bool {
	return strings.Contains(err.Error(), "NOT_FOUND")
}

// IsUnauthorized checks if the error is an "unauthorized" error
func IsUnauthorized(err error) bool {
	return strings.Contains(err.Error(), "UNAUTHORIZED")
}

func compactWhitespace(input string) string {
	return strings.Join(strings.Fields(input), " ")
}

func escapeMarkdown(input string) string {
	clean := compactWhitespace(input)
	return strings.ReplaceAll(clean, "|", "\\|")
}

func printFeedbackTable(resp *FeedbackResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Created\tEmail\tComment")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			item.Attributes.CreatedDate,
			item.Attributes.Email,
			compactWhitespace(item.Attributes.Comment),
		)
	}
	return w.Flush()
}

func printCrashesTable(resp *CrashesResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Created\tEmail\tDevice\tOS\tComment")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			item.Attributes.CreatedDate,
			item.Attributes.Email,
			item.Attributes.DeviceModel,
			item.Attributes.OSVersion,
			compactWhitespace(item.Attributes.Comment),
		)
	}
	return w.Flush()
}

func printReviewsTable(resp *ReviewsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Created\tRating\tTerritory\tTitle")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\n",
			item.Attributes.CreatedDate,
			item.Attributes.Rating,
			item.Attributes.Territory,
			compactWhitespace(item.Attributes.Title),
		)
	}
	return w.Flush()
}

func printFeedbackMarkdown(resp *FeedbackResponse) error {
	fmt.Fprintln(os.Stdout, "| Created | Email | Comment |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s |\n",
			escapeMarkdown(item.Attributes.CreatedDate),
			escapeMarkdown(item.Attributes.Email),
			escapeMarkdown(item.Attributes.Comment),
		)
	}
	return nil
}

func printCrashesMarkdown(resp *CrashesResponse) error {
	fmt.Fprintln(os.Stdout, "| Created | Email | Device | OS | Comment |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s |\n",
			escapeMarkdown(item.Attributes.CreatedDate),
			escapeMarkdown(item.Attributes.Email),
			escapeMarkdown(item.Attributes.DeviceModel),
			escapeMarkdown(item.Attributes.OSVersion),
			escapeMarkdown(item.Attributes.Comment),
		)
	}
	return nil
}

func printReviewsMarkdown(resp *ReviewsResponse) error {
	fmt.Fprintln(os.Stdout, "| Created | Rating | Territory | Title |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %d | %s | %s |\n",
			escapeMarkdown(item.Attributes.CreatedDate),
			item.Attributes.Rating,
			escapeMarkdown(item.Attributes.Territory),
			escapeMarkdown(item.Attributes.Title),
		)
	}
	return nil
}

func printAppsTable(resp *AppsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tBundle ID\tSKU")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			item.ID,
			compactWhitespace(item.Attributes.Name),
			item.Attributes.BundleID,
			item.Attributes.SKU,
		)
	}
	return w.Flush()
}

func printBuildsTable(resp *BuildsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Version\tUploaded\tProcessing\tExpired")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%t\n",
			item.Attributes.Version,
			item.Attributes.UploadedDate,
			item.Attributes.ProcessingState,
			item.Attributes.Expired,
		)
	}
	return w.Flush()
}

func printAppsMarkdown(resp *AppsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Name | Bundle ID | SKU |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s |\n",
			item.ID,
			escapeMarkdown(item.Attributes.Name),
			escapeMarkdown(item.Attributes.BundleID),
			escapeMarkdown(item.Attributes.SKU),
		)
	}
	return nil
}

func printBuildsMarkdown(resp *BuildsResponse) error {
	fmt.Fprintln(os.Stdout, "| Version | Uploaded | Processing | Expired |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %t |\n",
			escapeMarkdown(item.Attributes.Version),
			escapeMarkdown(item.Attributes.UploadedDate),
			escapeMarkdown(item.Attributes.ProcessingState),
			item.Attributes.Expired,
		)
	}
	return nil
}
