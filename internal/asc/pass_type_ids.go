package asc

// PassTypeIDAttributes describes a pass type ID resource.
type PassTypeIDAttributes struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

// PassTypeIDCreateAttributes describes attributes for creating a pass type ID.
type PassTypeIDCreateAttributes struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

// PassTypeIDUpdateAttributes describes attributes for updating a pass type ID.
type PassTypeIDUpdateAttributes struct {
	Name string `json:"name,omitempty"`
}

// PassTypeIDCreateData is the data portion of a pass type ID create request.
type PassTypeIDCreateData struct {
	Type       ResourceType               `json:"type"`
	Attributes PassTypeIDCreateAttributes `json:"attributes"`
}

// PassTypeIDCreateRequest is a request to create a pass type ID.
type PassTypeIDCreateRequest struct {
	Data PassTypeIDCreateData `json:"data"`
}

// PassTypeIDUpdateData is the data portion of a pass type ID update request.
type PassTypeIDUpdateData struct {
	Type       ResourceType                `json:"type"`
	ID         string                      `json:"id"`
	Attributes *PassTypeIDUpdateAttributes `json:"attributes,omitempty"`
}

// PassTypeIDUpdateRequest is a request to update a pass type ID.
type PassTypeIDUpdateRequest struct {
	Data PassTypeIDUpdateData `json:"data"`
}

// PassTypeIDsResponse is the response from pass type IDs list endpoint.
type PassTypeIDsResponse = Response[PassTypeIDAttributes]

// PassTypeIDResponse is the response from pass type ID detail endpoint.
type PassTypeIDResponse = SingleResponse[PassTypeIDAttributes]

// PassTypeIDCertificatesLinkagesResponse is the response from pass type ID certificates linkages endpoints.
type PassTypeIDCertificatesLinkagesResponse = LinkagesResponse

// CertificatePassTypeIDLinkageResponse is the response from certificate pass type ID linkage endpoint.
type CertificatePassTypeIDLinkageResponse struct {
	Data  ResourceData `json:"data"`
	Links Links        `json:"links,omitempty"`
}
