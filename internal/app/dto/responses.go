package dto

type ErrorResponse struct {
	Message   string `json:"message"`
	ErrorCode int    `json:"errorCode"`
}

type AuthenticateStatusResponse struct {
	IsAuthenticated bool `json:"authenticated"`
	StatusCode      int  `json:"statusCode"`
}

type SuccessResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}
