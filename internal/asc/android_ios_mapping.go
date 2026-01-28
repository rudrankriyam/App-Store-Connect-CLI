package asc

// AndroidToIosAppMappingDetailAttributes describes an android-to-iOS mapping resource.
type AndroidToIosAppMappingDetailAttributes struct {
	PackageName                                      string   `json:"packageName,omitempty"`
	AppSigningKeyPublicCertificateSha256Fingerprints []string `json:"appSigningKeyPublicCertificateSha256Fingerprints,omitempty"`
}

// AndroidToIosAppMappingDetailCreateAttributes describes attributes for creating a mapping.
type AndroidToIosAppMappingDetailCreateAttributes struct {
	PackageName                                      string   `json:"packageName"`
	AppSigningKeyPublicCertificateSha256Fingerprints []string `json:"appSigningKeyPublicCertificateSha256Fingerprints"`
}

// AndroidToIosAppMappingDetailUpdateAttributes describes attributes for updating a mapping.
type AndroidToIosAppMappingDetailUpdateAttributes struct {
	PackageName                                      string   `json:"packageName,omitempty"`
	AppSigningKeyPublicCertificateSha256Fingerprints []string `json:"appSigningKeyPublicCertificateSha256Fingerprints,omitempty"`
}

// AndroidToIosAppMappingDetailCreateRelationships describes relationships for creating a mapping.
type AndroidToIosAppMappingDetailCreateRelationships struct {
	App *Relationship `json:"app"`
}

// AndroidToIosAppMappingDetailCreateData is the data portion of a create request.
type AndroidToIosAppMappingDetailCreateData struct {
	Type          ResourceType                                    `json:"type"`
	Attributes    AndroidToIosAppMappingDetailCreateAttributes    `json:"attributes"`
	Relationships AndroidToIosAppMappingDetailCreateRelationships `json:"relationships"`
}

// AndroidToIosAppMappingDetailCreateRequest is a request to create a mapping.
type AndroidToIosAppMappingDetailCreateRequest struct {
	Data AndroidToIosAppMappingDetailCreateData `json:"data"`
}

// AndroidToIosAppMappingDetailUpdateData is the data portion of an update request.
type AndroidToIosAppMappingDetailUpdateData struct {
	Type       ResourceType                                  `json:"type"`
	ID         string                                        `json:"id"`
	Attributes *AndroidToIosAppMappingDetailUpdateAttributes `json:"attributes,omitempty"`
}

// AndroidToIosAppMappingDetailUpdateRequest is a request to update a mapping.
type AndroidToIosAppMappingDetailUpdateRequest struct {
	Data AndroidToIosAppMappingDetailUpdateData `json:"data"`
}

// AndroidToIosAppMappingDetailsResponse is the response from the mapping list endpoint.
type AndroidToIosAppMappingDetailsResponse = Response[AndroidToIosAppMappingDetailAttributes]

// AndroidToIosAppMappingDetailResponse is the response from the mapping detail endpoint.
type AndroidToIosAppMappingDetailResponse = SingleResponse[AndroidToIosAppMappingDetailAttributes]
