package promapi

// Result is a shared type for tool handlers that return a boolean success
// status with an optional error message. Used by manage (health, readiness,
// reload, quit) and tsdbadmin (delete-series, clean-tombstones) handlers.
type Result struct {
	Success bool   `json:"success" jsonschema:"Indicate the result of the operation, true means success, false means failure"`
	Message string `json:"message,omitempty" jsonschema:"Explanation message when the operation fails."`
}

// ResultOf returns a Result based on err. If err is nil, the result
// indicates success. Otherwise, the result indicates failure with the
// error message.
//
// ResultOf returns a Result based on err. If err is nil, the result
// indicates success. Otherwise, it indicates failure with the error message.
func ResultOf(err error) *Result {
	if err != nil {
		return &Result{Success: false, Message: err.Error()}
	}
	return &Result{Success: true}
}
