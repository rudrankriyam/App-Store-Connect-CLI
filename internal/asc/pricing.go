package asc

// TerritoryAttributes describes a territory resource.
type TerritoryAttributes struct {
	Currency string `json:"currency,omitempty"`
}

// AppPricePointV3Attributes describes app price point metadata.
type AppPricePointV3Attributes struct {
	CustomerPrice string `json:"customerPrice,omitempty"`
	Proceeds      string `json:"proceeds,omitempty"`
}

// AppPriceScheduleAttributes describes an app price schedule resource.
type AppPriceScheduleAttributes struct {
	// Usually empty - data is in relationships.
}

// AppPriceAttributes describes a price schedule entry.
type AppPriceAttributes struct {
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
	Manual    bool   `json:"manual,omitempty"`
}

// AppAvailabilityV2Attributes describes app availability metadata.
type AppAvailabilityV2Attributes struct {
	AvailableInNewTerritories bool `json:"availableInNewTerritories"`
}

// TerritoryAvailabilityAttributes describes availability for a territory.
type TerritoryAvailabilityAttributes struct {
	Available       bool   `json:"available"`
	ReleaseDate     string `json:"releaseDate,omitempty"`
	PreOrderEnabled bool   `json:"preOrderEnabled,omitempty"`
}

// Response types
type TerritoriesResponse = Response[TerritoryAttributes]
type AppPricePointsV3Response = Response[AppPricePointV3Attributes]
type AppPriceScheduleResponse = SingleResponse[AppPriceScheduleAttributes]
type AppPricesResponse = Response[AppPriceAttributes]
type AppAvailabilityV2Response = SingleResponse[AppAvailabilityV2Attributes]
type TerritoryAvailabilitiesResponse = Response[TerritoryAvailabilityAttributes]

// AppPriceScheduleCreateAttributes defines inputs for creating a price schedule.
type AppPriceScheduleCreateAttributes struct {
	PricePointID string `json:"-"`
	StartDate    string `json:"-"`
}

// AppPriceScheduleCreateRequest is a request to create a price schedule.
type AppPriceScheduleCreateRequest struct {
	Data     AppPriceScheduleCreateData `json:"data"`
	Included []AppPriceCreateResource   `json:"included,omitempty"`
}

// AppPriceScheduleCreateData is the data portion of a schedule create request.
type AppPriceScheduleCreateData struct {
	Type          ResourceType                        `json:"type"`
	Relationships AppPriceScheduleCreateRelationships `json:"relationships"`
}

// AppPriceScheduleCreateRelationships describes schedule relationships.
type AppPriceScheduleCreateRelationships struct {
	App          Relationship     `json:"app"`
	ManualPrices RelationshipList `json:"manualPrices,omitempty"`
}

// AppPriceCreateResource represents an app price resource for schedule creation.
type AppPriceCreateResource struct {
	Type          ResourceType          `json:"type"`
	ID            string                `json:"id,omitempty"`
	Attributes    AppPriceAttributes    `json:"attributes"`
	Relationships AppPriceRelationships `json:"relationships"`
}

// AppPriceRelationships describes relationships for app prices.
type AppPriceRelationships struct {
	AppPricePoint Relationship `json:"appPricePoint"`
}

// AppAvailabilityV2CreateAttributes defines inputs for app availability.
type AppAvailabilityV2CreateAttributes struct {
	AvailableInNewTerritories *bool                         `json:"availableInNewTerritories,omitempty"`
	TerritoryAvailabilities   []TerritoryAvailabilityCreate `json:"-"`
}

// TerritoryAvailabilityCreate defines a territory availability input.
type TerritoryAvailabilityCreate struct {
	TerritoryID     string
	Available       bool
	ReleaseDate     string
	PreOrderEnabled *bool
}

// AppAvailabilityV2CreateRequest is a request to create app availability.
type AppAvailabilityV2CreateRequest struct {
	Data     AppAvailabilityV2CreateData           `json:"data"`
	Included []TerritoryAvailabilityCreateResource `json:"included,omitempty"`
}

// AppAvailabilityV2CreateData is the data portion of availability create.
type AppAvailabilityV2CreateData struct {
	Type          ResourceType                         `json:"type"`
	Attributes    *AppAvailabilityV2CreateAttributes   `json:"attributes,omitempty"`
	Relationships AppAvailabilityV2CreateRelationships `json:"relationships"`
}

// AppAvailabilityV2CreateRelationships describes availability relationships.
type AppAvailabilityV2CreateRelationships struct {
	App                     Relationship     `json:"app"`
	TerritoryAvailabilities RelationshipList `json:"territoryAvailabilities,omitempty"`
}

// TerritoryAvailabilityCreateAttributes describes attributes for create.
type TerritoryAvailabilityCreateAttributes struct {
	Available       bool   `json:"available"`
	ReleaseDate     string `json:"releaseDate,omitempty"`
	PreOrderEnabled *bool  `json:"preOrderEnabled,omitempty"`
}

// TerritoryAvailabilityCreateResource represents a create payload resource.
type TerritoryAvailabilityCreateResource struct {
	Type          ResourceType                          `json:"type"`
	ID            string                                `json:"id,omitempty"`
	Attributes    TerritoryAvailabilityCreateAttributes `json:"attributes"`
	Relationships TerritoryAvailabilityRelationships    `json:"relationships"`
}

// TerritoryAvailabilityRelationships describes relationships for availability.
type TerritoryAvailabilityRelationships struct {
	Territory Relationship `json:"territory"`
}
