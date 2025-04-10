package core

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strconv"
)

// ResponseData is a generic struct for API responses
type ResponseData struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// PaginationMeta contains pagination information
type PaginationMeta struct {
	CurrentPage int  `json:"current_page"`
	PerPage     int  `json:"per_page"`
	TotalItems  int  `json:"total_items"`
	TotalPages  int  `json:"total_pages"`
	HasNextPage bool `json:"has_next_page"`
	HasPrevPage bool `json:"has_prev_page"`
}

// NewPaginationMeta creates a new pagination metadata object
func NewPaginationMeta(currentPage, perPage, totalItems int) PaginationMeta {
	totalPages := int(math.Ceil(float64(totalItems) / float64(perPage)))

	return PaginationMeta{
		CurrentPage: currentPage,
		PerPage:     perPage,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasNextPage: currentPage < totalPages,
		HasPrevPage: currentPage > 1,
	}
}

// GetPaginationParams extracts pagination parameters from request
func GetPaginationParams(r *http.Request, defaultPerPage int) (page, perPage int) {
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("per_page")

	page = 1
	perPage = defaultPerPage

	if pageStr != "" {
		if pageInt, err := strconv.Atoi(pageStr); err == nil && pageInt > 0 {
			page = pageInt
		}
	}

	if perPageStr != "" {
		if perPageInt, err := strconv.Atoi(perPageStr); err == nil && perPageInt > 0 {
			perPage = perPageInt
		}
	}

	return page, perPage
}

// RenderJSON marshals data to JSON and writes it to the response writer
func RenderJSON(w http.ResponseWriter, data interface{}, statusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	return json.NewEncoder(w).Encode(data)
}

// RenderSuccess is a convenience function to render a successful JSON response
func RenderSuccess(w http.ResponseWriter, data interface{}, statusCode int) error {
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	response := ResponseData{
		Success: true,
		Data:    data,
	}

	return RenderJSON(w, response, statusCode)
}

// RenderError is a convenience function to render an error JSON response
func RenderError(w http.ResponseWriter, errMessage string, statusCode int) error {
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}

	response := ResponseData{
		Success: false,
		Error:   errMessage,
	}

	return RenderJSON(w, response, statusCode)
}

// RenderPaginated creates a paginated JSON response
func RenderPaginated(w http.ResponseWriter, data interface{}, meta PaginationMeta, statusCode int) error {
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	response := ResponseData{
		Success: true,
		Data:    data,
		Meta:    meta,
	}

	return RenderJSON(w, response, statusCode)
}

// ParseBody parses the request body into the provided interface
func ParseBody(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return errors.New("request body is empty")
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	if len(body) == 0 {
		return errors.New("request body is empty")
	}

	return json.Unmarshal(body, v)
}

// ParseJSONParams parses URL query params into a map
func ParseJSONParams(r *http.Request) map[string]interface{} {
	params := make(map[string]interface{})

	query := r.URL.Query()
	for key, values := range query {
		if len(values) > 1 {
			params[key] = values
		} else if len(values) == 1 {
			params[key] = values[0]
		}
	}

	return params
}

// GetParam extracts a parameter from the request by name, with type conversion
func GetParam(r *http.Request, name string, defaultValue string) string {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetParamInt extracts an integer parameter from the request by name
func GetParamInt(r *http.Request, name string, defaultValue int) int {
	strValue := r.URL.Query().Get(name)
	if strValue == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(strValue)
	if err != nil {
		return defaultValue
	}

	return value
}

// GetParamBool extracts a boolean parameter from the request by name
func GetParamBool(r *http.Request, name string, defaultValue bool) bool {
	strValue := r.URL.Query().Get(name)
	if strValue == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(strValue)
	if err != nil {
		return defaultValue
	}

	return value
}

// IsJSONRequest checks if the request content type is JSON
func IsJSONRequest(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	return contentType == "application/json" || contentType == "application/json; charset=utf-8"
}

// APIResponse creates a standard API response
func APIResponse(success bool, data interface{}, errMessage string) ResponseData {
	return ResponseData{
		Success: success,
		Data:    data,
		Error:   errMessage,
	}
}

// WithMeta adds metadata to the response
func (r ResponseData) WithMeta(meta interface{}) ResponseData {
	r.Meta = meta
	return r
}
