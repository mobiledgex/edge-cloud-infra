package doc

// swagger:response success
type successResponse struct {
	// in: body
	Body struct {
		// HTTP status code 200 - Status OK
		Code int `json:"code,omitempty"`
		// Response data from controller
		Message string `json:"message,omitempty"`
	}
}

// swagger:response badRequest
type badReqResponse struct {
	// in: body
	Body struct {
		// HTTP status code 400 - Status Bad Request
		Code int `json:"code,omitempty"`
		// Detailed error message
		Message string `json:"message,omitempty"`
	}
}

// swagger:response forbidden
type forbiddenResponse struct {
	// in: body
	Body struct {
		// HTTP status code 403 - Forbidden
		Code int `json:"code,omitempty"`
		// Detailed error message
		Message string `json:"message,omitempty"`
	}
}

// swagger:response notFound
type notFoundResponse struct {
	// in: body
	Body struct {
		// HTTP status code 404 - Not Found
		Code int `json:"code,omitempty"`
		// Detailed error message
		Message string `json:"message,omitempty"`
	}
}
