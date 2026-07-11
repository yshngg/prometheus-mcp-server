package promapi

// Result is a shared type for tool handlers that return a boolean success
// status with an optional error message. Used by manage (health, readiness,
// reload, quit) and tsdbadmin (delete-series, clean-tombstones) handlers.
type Result struct {
	Success bool   `json:"success" jsonschema:"Indicate the result of the operation, true means success, false means failure"`
	Message string `json:"message,omitempty" jsonschema:"Explanation message when the operation fails."`
}
