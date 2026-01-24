package asc

// BetaRecruitmentCriteriaAttributes describes beta recruitment criteria metadata.
type BetaRecruitmentCriteriaAttributes struct {
	LastModifiedDate string `json:"lastModifiedDate,omitempty"`
}

// BetaRecruitmentCriteriaResponse is the response from beta recruitment criteria endpoints.
type BetaRecruitmentCriteriaResponse = SingleResponse[BetaRecruitmentCriteriaAttributes]

// BetaRecruitmentCriteriaRelationships describes relationships for recruitment criteria.
type BetaRecruitmentCriteriaRelationships struct {
	BetaGroup                       *Relationship     `json:"betaGroup,omitempty"`
	BetaRecruitmentCriterionOptions *RelationshipList `json:"betaRecruitmentCriterionOptions,omitempty"`
}

// BetaRecruitmentCriteriaCreateData is the data portion of a criteria create request.
type BetaRecruitmentCriteriaCreateData struct {
	Type          ResourceType                          `json:"type"`
	Relationships *BetaRecruitmentCriteriaRelationships `json:"relationships"`
}

// BetaRecruitmentCriteriaCreateRequest is a request to create beta recruitment criteria.
type BetaRecruitmentCriteriaCreateRequest struct {
	Data BetaRecruitmentCriteriaCreateData `json:"data"`
}

// BetaRecruitmentCriteriaUpdateData is the data portion of a criteria update request.
type BetaRecruitmentCriteriaUpdateData struct {
	Type          ResourceType                          `json:"type"`
	ID            string                                `json:"id"`
	Relationships *BetaRecruitmentCriteriaRelationships `json:"relationships,omitempty"`
}

// BetaRecruitmentCriteriaUpdateRequest is a request to update beta recruitment criteria.
type BetaRecruitmentCriteriaUpdateRequest struct {
	Data BetaRecruitmentCriteriaUpdateData `json:"data"`
}

// BetaRecruitmentCriterionOptionAttributes describes recruitment criteria options.
type BetaRecruitmentCriterionOptionAttributes struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	Category   string `json:"category,omitempty"`
}

// BetaRecruitmentCriterionOptionsResponse is the response from recruitment criteria options list.
type BetaRecruitmentCriterionOptionsResponse = Response[BetaRecruitmentCriterionOptionAttributes]

// BetaGroupMetricAttributes represents metric attributes returned by metrics endpoints.
type BetaGroupMetricAttributes map[string]interface{}

// BetaGroupPublicLinkUsagesResponse is the response from public link usage metrics.
type BetaGroupPublicLinkUsagesResponse = Response[BetaGroupMetricAttributes]

// BetaGroupTesterUsagesResponse is the response from beta tester usage metrics.
type BetaGroupTesterUsagesResponse = Response[BetaGroupMetricAttributes]
