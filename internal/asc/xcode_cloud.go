package asc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// Xcode Cloud Resource Types
const (
	ResourceTypeCiProducts       ResourceType = "ciProducts"
	ResourceTypeCiWorkflows      ResourceType = "ciWorkflows"
	ResourceTypeCiBuildRuns      ResourceType = "ciBuildRuns"
	ResourceTypeCiBuildActions   ResourceType = "ciBuildActions"
	ResourceTypeScmRepositories  ResourceType = "scmRepositories"
	ResourceTypeScmGitReferences ResourceType = "scmGitReferences"
)

// CiBuildRunExecutionProgress represents the execution progress of a build run.
type CiBuildRunExecutionProgress string

const (
	CiBuildRunExecutionProgressPending  CiBuildRunExecutionProgress = "PENDING"
	CiBuildRunExecutionProgressRunning  CiBuildRunExecutionProgress = "RUNNING"
	CiBuildRunExecutionProgressComplete CiBuildRunExecutionProgress = "COMPLETE"
)

// CiBuildRunCompletionStatus represents the completion status of a build run.
type CiBuildRunCompletionStatus string

const (
	CiBuildRunCompletionStatusSucceeded CiBuildRunCompletionStatus = "SUCCEEDED"
	CiBuildRunCompletionStatusFailed    CiBuildRunCompletionStatus = "FAILED"
	CiBuildRunCompletionStatusErrored   CiBuildRunCompletionStatus = "ERRORED"
	CiBuildRunCompletionStatusCanceled  CiBuildRunCompletionStatus = "CANCELED"
	CiBuildRunCompletionStatusSkipped   CiBuildRunCompletionStatus = "SKIPPED"
)

// CiProductAttributes describes a CI product resource.
type CiProductAttributes struct {
	Name        string `json:"name,omitempty"`
	CreatedDate string `json:"createdDate,omitempty"`
	ProductType string `json:"productType,omitempty"`
	BundleID    string `json:"bundleId,omitempty"`
}

// CiProductRelationships describes relationships for a CI product.
type CiProductRelationships struct {
	App                 *Relationship     `json:"app,omitempty"`
	PrimaryRepositories *RelationshipList `json:"primaryRepositories,omitempty"`
}

// CiProductResource represents a CI product resource.
type CiProductResource struct {
	Type          ResourceType            `json:"type"`
	ID            string                  `json:"id"`
	Attributes    CiProductAttributes     `json:"attributes,omitempty"`
	Relationships *CiProductRelationships `json:"relationships,omitempty"`
}

// CiProductsResponse is the response from CI products endpoints.
type CiProductsResponse struct {
	Data  []CiProductResource `json:"data"`
	Links Links               `json:"links,omitempty"`
}

// GetLinks returns the links field for pagination.
func (r *CiProductsResponse) GetLinks() *Links {
	return &r.Links
}

// GetData returns the data field for aggregation.
func (r *CiProductsResponse) GetData() interface{} {
	return r.Data
}

// CiProductResponse is the response from CI product detail endpoints.
type CiProductResponse struct {
	Data  CiProductResource `json:"data"`
	Links Links             `json:"links,omitempty"`
}

// CiWorkflowAttributes describes a CI workflow resource.
type CiWorkflowAttributes struct {
	Name                            string                       `json:"name,omitempty"`
	Description                     string                       `json:"description,omitempty"`
	BranchStartCondition            *CiBranchStartCondition      `json:"branchStartCondition,omitempty"`
	TagStartCondition               *CiTagStartCondition         `json:"tagStartCondition,omitempty"`
	PullRequestStartCondition       *CiPullRequestStartCondition `json:"pullRequestStartCondition,omitempty"`
	ScheduledStartCondition         *CiScheduledStartCondition   `json:"scheduledStartCondition,omitempty"`
	ManualBranchStartCondition      *CiManualStartCondition      `json:"manualBranchStartCondition,omitempty"`
	ManualTagStartCondition         *CiManualStartCondition      `json:"manualTagStartCondition,omitempty"`
	ManualPullRequestStartCondition *CiManualStartCondition      `json:"manualPullRequestStartCondition,omitempty"`
	IsEnabled                       bool                         `json:"isEnabled,omitempty"`
	IsLockedForEditing              bool                         `json:"isLockedForEditing,omitempty"`
	Clean                           bool                         `json:"clean,omitempty"`
	ContainerFilePath               string                       `json:"containerFilePath,omitempty"`
	LastModifiedDate                string                       `json:"lastModifiedDate,omitempty"`
}

// CiBranchStartCondition describes branch start conditions.
type CiBranchStartCondition struct {
	Source              *CiBranchPatterns      `json:"source,omitempty"`
	FilesAndFoldersRule *CiFilesAndFoldersRule `json:"filesAndFoldersRule,omitempty"`
	AutoCancel          bool                   `json:"autoCancel,omitempty"`
}

// CiTagStartCondition describes tag start conditions.
type CiTagStartCondition struct {
	Source              *CiTagPatterns         `json:"source,omitempty"`
	FilesAndFoldersRule *CiFilesAndFoldersRule `json:"filesAndFoldersRule,omitempty"`
	AutoCancel          bool                   `json:"autoCancel,omitempty"`
}

// CiPullRequestStartCondition describes pull request start conditions.
type CiPullRequestStartCondition struct {
	Source              *CiBranchPatterns      `json:"source,omitempty"`
	Destination         *CiBranchPatterns      `json:"destination,omitempty"`
	FilesAndFoldersRule *CiFilesAndFoldersRule `json:"filesAndFoldersRule,omitempty"`
	AutoCancel          bool                   `json:"autoCancel,omitempty"`
}

// CiScheduledStartCondition describes scheduled start conditions.
type CiScheduledStartCondition struct {
	Source   *CiBranchPatterns `json:"source,omitempty"`
	Schedule *CiSchedule       `json:"schedule,omitempty"`
}

// CiManualStartCondition describes manual start conditions.
type CiManualStartCondition struct {
	Source *CiBranchPatterns `json:"source,omitempty"`
}

// CiBranchPatterns describes branch patterns.
type CiBranchPatterns struct {
	Patterns   []CiStartConditionPattern `json:"patterns,omitempty"`
	IsAllMatch bool                      `json:"isAllMatch,omitempty"`
}

// CiTagPatterns describes tag patterns.
type CiTagPatterns struct {
	Patterns   []CiStartConditionPattern `json:"patterns,omitempty"`
	IsAllMatch bool                      `json:"isAllMatch,omitempty"`
}

// CiStartConditionPattern describes a start condition pattern.
type CiStartConditionPattern struct {
	Pattern  string `json:"pattern,omitempty"`
	IsPrefix bool   `json:"isPrefix,omitempty"`
}

// CiFilesAndFoldersRule describes files and folders rules.
type CiFilesAndFoldersRule struct {
	Mode  string   `json:"mode,omitempty"`
	Paths []string `json:"paths,omitempty"`
}

// CiSchedule describes a CI schedule.
type CiSchedule struct {
	Frequency string   `json:"frequency,omitempty"`
	Days      []string `json:"days,omitempty"`
	Hour      int      `json:"hour,omitempty"`
	Minute    int      `json:"minute,omitempty"`
	Timezone  string   `json:"timezone,omitempty"`
}

// CiWorkflowRelationships describes relationships for a CI workflow.
type CiWorkflowRelationships struct {
	Product      *Relationship `json:"product,omitempty"`
	Repository   *Relationship `json:"repository,omitempty"`
	XcodeVersion *Relationship `json:"xcodeVersion,omitempty"`
	MacOsVersion *Relationship `json:"macOsVersion,omitempty"`
}

// CiWorkflowResource represents a CI workflow resource.
type CiWorkflowResource struct {
	Type          ResourceType             `json:"type"`
	ID            string                   `json:"id"`
	Attributes    CiWorkflowAttributes     `json:"attributes,omitempty"`
	Relationships *CiWorkflowRelationships `json:"relationships,omitempty"`
}

// CiWorkflowsResponse is the response from CI workflows endpoints.
type CiWorkflowsResponse struct {
	Data  []CiWorkflowResource `json:"data"`
	Links Links                `json:"links,omitempty"`
}

// GetLinks returns the links field for pagination.
func (r *CiWorkflowsResponse) GetLinks() *Links {
	return &r.Links
}

// GetData returns the data field for aggregation.
func (r *CiWorkflowsResponse) GetData() interface{} {
	return r.Data
}

// CiWorkflowResponse is the response from CI workflow detail endpoints.
type CiWorkflowResponse struct {
	Data  CiWorkflowResource `json:"data"`
	Links Links              `json:"links,omitempty"`
}

// ScmRepositoryAttributes describes an SCM repository resource.
type ScmRepositoryAttributes struct {
	HTTPCloneURL     string `json:"httpCloneUrl,omitempty"`
	SSHCloneURL      string `json:"sshCloneUrl,omitempty"`
	OwnerName        string `json:"ownerName,omitempty"`
	RepositoryName   string `json:"repositoryName,omitempty"`
	LastAccessedDate string `json:"lastAccessedDate,omitempty"`
}

// ScmRepositoryRelationships describes relationships for an SCM repository.
type ScmRepositoryRelationships struct {
	ScmProvider   *Relationship `json:"scmProvider,omitempty"`
	DefaultBranch *Relationship `json:"defaultBranch,omitempty"`
}

// ScmRepositoryResource represents an SCM repository resource.
type ScmRepositoryResource struct {
	Type          ResourceType                `json:"type"`
	ID            string                      `json:"id"`
	Attributes    ScmRepositoryAttributes     `json:"attributes,omitempty"`
	Relationships *ScmRepositoryRelationships `json:"relationships,omitempty"`
}

// ScmRepositoriesResponse is the response from SCM repositories endpoints.
type ScmRepositoriesResponse struct {
	Data  []ScmRepositoryResource `json:"data"`
	Links Links                   `json:"links,omitempty"`
}

// ScmGitReferenceAttributes describes an SCM git reference resource.
type ScmGitReferenceAttributes struct {
	Name          string `json:"name,omitempty"`
	CanonicalName string `json:"canonicalName,omitempty"`
	IsDeleted     bool   `json:"isDeleted,omitempty"`
	Kind          string `json:"kind,omitempty"` // BRANCH or TAG
}

// ScmGitReferenceRelationships describes relationships for an SCM git reference.
type ScmGitReferenceRelationships struct {
	Repository *Relationship `json:"repository,omitempty"`
}

// ScmGitReferenceResource represents an SCM git reference resource.
type ScmGitReferenceResource struct {
	Type          ResourceType                  `json:"type"`
	ID            string                        `json:"id"`
	Attributes    ScmGitReferenceAttributes     `json:"attributes,omitempty"`
	Relationships *ScmGitReferenceRelationships `json:"relationships,omitempty"`
}

// ScmGitReferencesResponse is the response from SCM git references endpoints.
type ScmGitReferencesResponse struct {
	Data  []ScmGitReferenceResource `json:"data"`
	Links Links                     `json:"links,omitempty"`
}

// GetLinks returns the links field for pagination.
func (r *ScmGitReferencesResponse) GetLinks() *Links {
	return &r.Links
}

// GetData returns the data field for aggregation.
func (r *ScmGitReferencesResponse) GetData() interface{} {
	return r.Data
}

// CiBuildRunAttributes describes a CI build run resource.
type CiBuildRunAttributes struct {
	Number             int                         `json:"number,omitempty"`
	CreatedDate        string                      `json:"createdDate,omitempty"`
	StartedDate        string                      `json:"startedDate,omitempty"`
	FinishedDate       string                      `json:"finishedDate,omitempty"`
	SourceCommit       *CiGitRefInfo               `json:"sourceCommit,omitempty"`
	DestinationCommit  *CiGitRefInfo               `json:"destinationCommit,omitempty"`
	IsPullRequestBuild bool                        `json:"isPullRequestBuild,omitempty"`
	IssueCounts        *CiIssueCounts              `json:"issueCounts,omitempty"`
	ExecutionProgress  CiBuildRunExecutionProgress `json:"executionProgress,omitempty"`
	CompletionStatus   CiBuildRunCompletionStatus  `json:"completionStatus,omitempty"`
	StartReason        string                      `json:"startReason,omitempty"`
	CancelReason       string                      `json:"cancelReason,omitempty"`
}

// CiGitRefInfo describes git reference information.
type CiGitRefInfo struct {
	CommitSha string     `json:"commitSha,omitempty"`
	Author    *CiGitUser `json:"author,omitempty"`
	Committer *CiGitUser `json:"committer,omitempty"`
	Message   string     `json:"message,omitempty"`
	WebURL    string     `json:"webUrl,omitempty"`
}

// CiGitUser describes a git user.
type CiGitUser struct {
	DisplayName string `json:"displayName,omitempty"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
}

// CiIssueCounts describes issue counts.
type CiIssueCounts struct {
	AnalyzerWarnings int `json:"analyzerWarnings,omitempty"`
	Errors           int `json:"errors,omitempty"`
	TestFailures     int `json:"testFailures,omitempty"`
	Warnings         int `json:"warnings,omitempty"`
}

// CiBuildRunRelationships describes relationships for a CI build run.
type CiBuildRunRelationships struct {
	Builds            *RelationshipList `json:"builds,omitempty"`
	Workflow          *Relationship     `json:"workflow,omitempty"`
	Product           *Relationship     `json:"product,omitempty"`
	SourceBranchOrTag *Relationship     `json:"sourceBranchOrTag,omitempty"`
	DestinationBranch *Relationship     `json:"destinationBranch,omitempty"`
	PullRequest       *Relationship     `json:"pullRequest,omitempty"`
}

// CiBuildRunResource represents a CI build run resource.
type CiBuildRunResource struct {
	Type          ResourceType             `json:"type"`
	ID            string                   `json:"id"`
	Attributes    CiBuildRunAttributes     `json:"attributes,omitempty"`
	Relationships *CiBuildRunRelationships `json:"relationships,omitempty"`
}

// CiBuildRunsResponse is the response from CI build runs endpoints.
type CiBuildRunsResponse struct {
	Data  []CiBuildRunResource `json:"data"`
	Links Links                `json:"links,omitempty"`
}

// GetLinks returns the links field for pagination.
func (r *CiBuildRunsResponse) GetLinks() *Links {
	return &r.Links
}

// GetData returns the data field for aggregation.
func (r *CiBuildRunsResponse) GetData() interface{} {
	return r.Data
}

// CiBuildRunResponse is the response from CI build run detail/create endpoints.
type CiBuildRunResponse struct {
	Data  CiBuildRunResource `json:"data"`
	Links Links              `json:"links,omitempty"`
}

// CiBuildRunCreateRequest is a request to create a CI build run.
type CiBuildRunCreateRequest struct {
	Data CiBuildRunCreateData `json:"data"`
}

// CiBuildRunCreateData is the data portion of a CI build run create request.
type CiBuildRunCreateData struct {
	Type          ResourceType                   `json:"type"`
	Relationships *CiBuildRunCreateRelationships `json:"relationships"`
}

// CiBuildRunCreateRelationships describes relationships for creating a CI build run.
type CiBuildRunCreateRelationships struct {
	Workflow          *Relationship `json:"workflow"`
	SourceBranchOrTag *Relationship `json:"sourceBranchOrTag"`
}

// Query types for Xcode Cloud endpoints

type ciProductsQuery struct {
	listQuery
	appID string
}

// CiProductsOption is a functional option for GetCiProducts.
type CiProductsOption func(*ciProductsQuery)

// WithCiProductsLimit sets the max number of CI products to return.
func WithCiProductsLimit(limit int) CiProductsOption {
	return func(q *ciProductsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithCiProductsNextURL uses a next page URL directly.
func WithCiProductsNextURL(next string) CiProductsOption {
	return func(q *ciProductsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithCiProductsAppID filters CI products by app ID.
func WithCiProductsAppID(appID string) CiProductsOption {
	return func(q *ciProductsQuery) {
		if strings.TrimSpace(appID) != "" {
			q.appID = strings.TrimSpace(appID)
		}
	}
}

func buildCiProductsQuery(query *ciProductsQuery) string {
	values := url.Values{}
	if query.appID != "" {
		values.Set("filter[app]", query.appID)
	}
	addLimit(values, query.limit)
	return values.Encode()
}

type ciWorkflowsQuery struct {
	listQuery
}

// CiWorkflowsOption is a functional option for GetCiWorkflows.
type CiWorkflowsOption func(*ciWorkflowsQuery)

// WithCiWorkflowsLimit sets the max number of CI workflows to return.
func WithCiWorkflowsLimit(limit int) CiWorkflowsOption {
	return func(q *ciWorkflowsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithCiWorkflowsNextURL uses a next page URL directly.
func WithCiWorkflowsNextURL(next string) CiWorkflowsOption {
	return func(q *ciWorkflowsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

func buildCiWorkflowsQuery(query *ciWorkflowsQuery) string {
	values := url.Values{}
	addLimit(values, query.limit)
	return values.Encode()
}

type scmGitReferencesQuery struct {
	listQuery
}

// ScmGitReferencesOption is a functional option for GetScmGitReferences.
type ScmGitReferencesOption func(*scmGitReferencesQuery)

// WithScmGitReferencesLimit sets the max number of git references to return.
func WithScmGitReferencesLimit(limit int) ScmGitReferencesOption {
	return func(q *scmGitReferencesQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithScmGitReferencesNextURL uses a next page URL directly.
func WithScmGitReferencesNextURL(next string) ScmGitReferencesOption {
	return func(q *scmGitReferencesQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

func buildScmGitReferencesQuery(query *scmGitReferencesQuery) string {
	values := url.Values{}
	addLimit(values, query.limit)
	return values.Encode()
}

type ciBuildRunsQuery struct {
	listQuery
}

// CiBuildRunsOption is a functional option for GetCiBuildRuns.
type CiBuildRunsOption func(*ciBuildRunsQuery)

// WithCiBuildRunsLimit sets the max number of build runs to return.
func WithCiBuildRunsLimit(limit int) CiBuildRunsOption {
	return func(q *ciBuildRunsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithCiBuildRunsNextURL uses a next page URL directly.
func WithCiBuildRunsNextURL(next string) CiBuildRunsOption {
	return func(q *ciBuildRunsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

func buildCiBuildRunsQuery(query *ciBuildRunsQuery) string {
	values := url.Values{}
	addLimit(values, query.limit)
	return values.Encode()
}

// GetCiProducts retrieves CI products, optionally filtered by app ID.
func (c *Client) GetCiProducts(ctx context.Context, opts ...CiProductsOption) (*CiProductsResponse, error) {
	query := &ciProductsQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := "/v1/ciProducts"
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("ciProducts: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildCiProductsQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response CiProductsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetCiProduct retrieves a CI product by ID.
func (c *Client) GetCiProduct(ctx context.Context, productID string) (*CiProductResponse, error) {
	path := fmt.Sprintf("/v1/ciProducts/%s", productID)
	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response CiProductResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetCiWorkflows retrieves CI workflows for a product.
func (c *Client) GetCiWorkflows(ctx context.Context, productID string, opts ...CiWorkflowsOption) (*CiWorkflowsResponse, error) {
	query := &ciWorkflowsQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/ciProducts/%s/workflows", productID)
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("ciWorkflows: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildCiWorkflowsQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response CiWorkflowsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetCiWorkflow retrieves a CI workflow by ID.
func (c *Client) GetCiWorkflow(ctx context.Context, workflowID string) (*CiWorkflowResponse, error) {
	path := fmt.Sprintf("/v1/ciWorkflows/%s", workflowID)
	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response CiWorkflowResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetCiWorkflowRepository retrieves the repository for a CI workflow.
func (c *Client) GetCiWorkflowRepository(ctx context.Context, workflowID string) (*ScmRepositoryResource, error) {
	path := fmt.Sprintf("/v1/ciWorkflows/%s/repository", workflowID)
	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data ScmRepositoryResource `json:"data"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response.Data, nil
}

// GetScmRepository retrieves an SCM repository by ID.
func (c *Client) GetScmRepository(ctx context.Context, repositoryID string) (*ScmRepositoryResource, error) {
	path := fmt.Sprintf("/v1/scmRepositories/%s", repositoryID)
	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data ScmRepositoryResource `json:"data"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response.Data, nil
}

// GetScmGitReferences retrieves git references for a repository.
func (c *Client) GetScmGitReferences(ctx context.Context, repositoryID string, opts ...ScmGitReferencesOption) (*ScmGitReferencesResponse, error) {
	query := &scmGitReferencesQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/scmRepositories/%s/gitReferences", repositoryID)
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("scmGitReferences: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildScmGitReferencesQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response ScmGitReferencesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetCiBuildRuns retrieves build runs for a workflow.
func (c *Client) GetCiBuildRuns(ctx context.Context, workflowID string, opts ...CiBuildRunsOption) (*CiBuildRunsResponse, error) {
	query := &ciBuildRunsQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/ciWorkflows/%s/buildRuns", workflowID)
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("ciBuildRuns: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildCiBuildRunsQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response CiBuildRunsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetCiBuildRun retrieves a CI build run by ID.
func (c *Client) GetCiBuildRun(ctx context.Context, buildRunID string) (*CiBuildRunResponse, error) {
	path := fmt.Sprintf("/v1/ciBuildRuns/%s", buildRunID)
	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response CiBuildRunResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// CreateCiBuildRun creates a new CI build run (triggers a workflow).
func (c *Client) CreateCiBuildRun(ctx context.Context, req CiBuildRunCreateRequest) (*CiBuildRunResponse, error) {
	body, err := BuildRequestBody(req)
	if err != nil {
		return nil, err
	}

	data, err := c.do(ctx, "POST", "/v1/ciBuildRuns", body)
	if err != nil {
		return nil, err
	}

	var response CiBuildRunResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// ResolveCiProductForApp finds the CI product for a given app ID.
// Returns an error if no product or multiple products are found.
func (c *Client) ResolveCiProductForApp(ctx context.Context, appID string) (*CiProductResource, error) {
	resp, err := c.GetCiProducts(ctx, WithCiProductsAppID(appID), WithCiProductsLimit(200))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CI products: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no Xcode Cloud product found for app %q (ensure Xcode Cloud is enabled)", appID)
	}

	if len(resp.Data) > 1 {
		return nil, fmt.Errorf("multiple Xcode Cloud products found for app %q; this is unexpected", appID)
	}

	return &resp.Data[0], nil
}

// ResolveCiWorkflowByName finds a workflow by name for a given product.
// Returns an error if no workflow or multiple workflows match the name.
func (c *Client) ResolveCiWorkflowByName(ctx context.Context, productID, workflowName string) (*CiWorkflowResource, error) {
	var allWorkflows []CiWorkflowResource
	var nextURL string

	for {
		var resp *CiWorkflowsResponse
		var err error

		if nextURL != "" {
			resp, err = c.GetCiWorkflows(ctx, productID, WithCiWorkflowsNextURL(nextURL))
		} else {
			resp, err = c.GetCiWorkflows(ctx, productID, WithCiWorkflowsLimit(200))
		}
		if err != nil {
			return nil, fmt.Errorf("failed to fetch CI workflows: %w", err)
		}

		allWorkflows = append(allWorkflows, resp.Data...)

		if resp.Links.Next == "" {
			break
		}
		nextURL = resp.Links.Next
	}

	if len(allWorkflows) == 0 {
		return nil, fmt.Errorf("no Xcode Cloud workflows found for product %q", productID)
	}

	// Find matching workflows by name (case-insensitive)
	var matches []CiWorkflowResource
	normalizedName := strings.ToLower(strings.TrimSpace(workflowName))
	for _, wf := range allWorkflows {
		if strings.ToLower(wf.Attributes.Name) == normalizedName {
			matches = append(matches, wf)
		}
	}

	if len(matches) == 0 {
		// List available workflows in error message
		var names []string
		for _, wf := range allWorkflows {
			names = append(names, wf.Attributes.Name)
		}
		return nil, fmt.Errorf("no workflow named %q found; available: %s", workflowName, strings.Join(names, ", "))
	}

	if len(matches) > 1 {
		var ids []string
		for _, wf := range matches {
			ids = append(ids, wf.ID)
		}
		return nil, fmt.Errorf("multiple workflows named %q found; use --workflow-id with one of: %s", workflowName, strings.Join(ids, ", "))
	}

	return &matches[0], nil
}

// ResolveGitReferenceByName finds a git reference (branch or tag) by name.
// Returns an error if no reference or multiple references match the name.
func (c *Client) ResolveGitReferenceByName(ctx context.Context, repositoryID, refName string) (*ScmGitReferenceResource, error) {
	var allRefs []ScmGitReferenceResource
	var nextURL string

	for {
		var resp *ScmGitReferencesResponse
		var err error

		if nextURL != "" {
			resp, err = c.GetScmGitReferences(ctx, repositoryID, WithScmGitReferencesNextURL(nextURL))
		} else {
			resp, err = c.GetScmGitReferences(ctx, repositoryID, WithScmGitReferencesLimit(200))
		}
		if err != nil {
			return nil, fmt.Errorf("failed to fetch git references: %w", err)
		}

		allRefs = append(allRefs, resp.Data...)

		if resp.Links.Next == "" {
			break
		}
		nextURL = resp.Links.Next
	}

	if len(allRefs) == 0 {
		return nil, fmt.Errorf("no git references found for repository %q", repositoryID)
	}

	// Find matching references by name
	normalizedName := strings.TrimSpace(refName)
	headsName := "refs/heads/" + normalizedName
	tagsName := "refs/tags/" + normalizedName
	var matches []ScmGitReferenceResource
	for _, ref := range allRefs {
		// Match by exact name or canonical ref (e.g., "main" or "refs/heads/main").
		canonical := ref.Attributes.CanonicalName
		if ref.Attributes.Name == normalizedName ||
			canonical == normalizedName ||
			canonical == headsName ||
			canonical == tagsName {
			if !ref.Attributes.IsDeleted {
				matches = append(matches, ref)
			}
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no git reference named %q found; use --git-reference-id to specify directly", refName)
	}

	if len(matches) > 1 {
		var ids []string
		for _, ref := range matches {
			ids = append(ids, fmt.Sprintf("%s (%s)", ref.ID, ref.Attributes.CanonicalName))
		}
		return nil, fmt.Errorf("multiple git references match %q; use --git-reference-id with one of: %s", refName, strings.Join(ids, ", "))
	}

	return &matches[0], nil
}

// XcodeCloudRunResult represents the result of triggering a build run.
type XcodeCloudRunResult struct {
	BuildRunID        string `json:"buildRunId"`
	BuildNumber       int    `json:"buildNumber,omitempty"`
	WorkflowID        string `json:"workflowId"`
	WorkflowName      string `json:"workflowName,omitempty"`
	GitReferenceID    string `json:"gitReferenceId"`
	GitReferenceName  string `json:"gitReferenceName,omitempty"`
	ExecutionProgress string `json:"executionProgress,omitempty"`
	CompletionStatus  string `json:"completionStatus,omitempty"`
	StartReason       string `json:"startReason,omitempty"`
	CreatedDate       string `json:"createdDate,omitempty"`
	StartedDate       string `json:"startedDate,omitempty"`
	FinishedDate      string `json:"finishedDate,omitempty"`
}

// XcodeCloudStatusResult represents the status of a build run.
type XcodeCloudStatusResult struct {
	BuildRunID        string         `json:"buildRunId"`
	BuildNumber       int            `json:"buildNumber,omitempty"`
	WorkflowID        string         `json:"workflowId,omitempty"`
	ExecutionProgress string         `json:"executionProgress"`
	CompletionStatus  string         `json:"completionStatus,omitempty"`
	StartReason       string         `json:"startReason,omitempty"`
	CancelReason      string         `json:"cancelReason,omitempty"`
	CreatedDate       string         `json:"createdDate,omitempty"`
	StartedDate       string         `json:"startedDate,omitempty"`
	FinishedDate      string         `json:"finishedDate,omitempty"`
	SourceCommit      *CiGitRefInfo  `json:"sourceCommit,omitempty"`
	IssueCounts       *CiIssueCounts `json:"issueCounts,omitempty"`
}

// IsBuildRunComplete returns true if the build run has finished.
func IsBuildRunComplete(progress CiBuildRunExecutionProgress) bool {
	return progress == CiBuildRunExecutionProgressComplete
}

// IsBuildRunSuccessful returns true if the build run completed successfully.
func IsBuildRunSuccessful(status CiBuildRunCompletionStatus) bool {
	return status == CiBuildRunCompletionStatusSucceeded
}

// CiBuildActionAttributes describes a CI build action resource.
type CiBuildActionAttributes struct {
	Name              string                      `json:"name,omitempty"`
	ActionType        string                      `json:"actionType,omitempty"` // BUILD, ANALYZE, TEST, ARCHIVE
	ExecutionProgress CiBuildRunExecutionProgress `json:"executionProgress,omitempty"`
	CompletionStatus  CiBuildRunCompletionStatus  `json:"completionStatus,omitempty"`
	StartedDate       string                      `json:"startedDate,omitempty"`
	FinishedDate      string                      `json:"finishedDate,omitempty"`
	IssueCounts       *CiIssueCounts              `json:"issueCounts,omitempty"`
}

// CiBuildActionResource represents a CI build action resource.
type CiBuildActionResource struct {
	Type       ResourceType            `json:"type"`
	ID         string                  `json:"id"`
	Attributes CiBuildActionAttributes `json:"attributes,omitempty"`
}

// CiBuildActionsResponse is the response from CI build actions endpoints.
type CiBuildActionsResponse struct {
	Data  []CiBuildActionResource `json:"data"`
	Links Links                   `json:"links,omitempty"`
}

// GetLinks returns the links field for pagination.
func (r *CiBuildActionsResponse) GetLinks() *Links {
	return &r.Links
}

// GetData returns the data field for aggregation.
func (r *CiBuildActionsResponse) GetData() interface{} {
	return r.Data
}

type ciBuildActionsQuery struct {
	listQuery
}

// CiBuildActionsOption is a functional option for GetCiBuildActions.
type CiBuildActionsOption func(*ciBuildActionsQuery)

// WithCiBuildActionsLimit sets the max number of build actions to return.
func WithCiBuildActionsLimit(limit int) CiBuildActionsOption {
	return func(q *ciBuildActionsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithCiBuildActionsNextURL uses a next page URL directly.
func WithCiBuildActionsNextURL(next string) CiBuildActionsOption {
	return func(q *ciBuildActionsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

func buildCiBuildActionsQuery(query *ciBuildActionsQuery) string {
	values := url.Values{}
	addLimit(values, query.limit)
	return values.Encode()
}

// GetCiBuildActions retrieves build actions for a build run.
func (c *Client) GetCiBuildActions(ctx context.Context, buildRunID string, opts ...CiBuildActionsOption) (*CiBuildActionsResponse, error) {
	query := &ciBuildActionsQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/ciBuildRuns/%s/actions", buildRunID)
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("ciBuildActions: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildCiBuildActionsQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response CiBuildActionsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}
