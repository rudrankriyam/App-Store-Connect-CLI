package asc

// MerchantIDAttributes describes a merchant ID resource.
type MerchantIDAttributes struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

// MerchantIDCreateAttributes describes attributes for creating a merchant ID.
type MerchantIDCreateAttributes struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

// MerchantIDUpdateAttributes describes attributes for updating a merchant ID.
type MerchantIDUpdateAttributes struct {
	Name string `json:"name,omitempty"`
}

// MerchantIDCreateData is the data portion of a merchant ID create request.
type MerchantIDCreateData struct {
	Type       ResourceType               `json:"type"`
	Attributes MerchantIDCreateAttributes `json:"attributes"`
}

// MerchantIDCreateRequest is a request to create a merchant ID.
type MerchantIDCreateRequest struct {
	Data MerchantIDCreateData `json:"data"`
}

// MerchantIDUpdateData is the data portion of a merchant ID update request.
type MerchantIDUpdateData struct {
	Type       ResourceType                `json:"type"`
	ID         string                      `json:"id"`
	Attributes *MerchantIDUpdateAttributes `json:"attributes,omitempty"`
}

// MerchantIDUpdateRequest is a request to update a merchant ID.
type MerchantIDUpdateRequest struct {
	Data MerchantIDUpdateData `json:"data"`
}

// MerchantIDsResponse is the response from merchant IDs list endpoint.
type MerchantIDsResponse = Response[MerchantIDAttributes]

// MerchantIDResponse is the response from merchant ID detail endpoint.
type MerchantIDResponse = SingleResponse[MerchantIDAttributes]

// MerchantIDCertificatesLinkagesResponse is the response from merchant ID certificate linkages endpoints.
type MerchantIDCertificatesLinkagesResponse = LinkagesResponse
