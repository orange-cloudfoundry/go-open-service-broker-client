package v2

// IsAsync returns true if the update request is being handled asynchronously.
func (r *UpdateInstanceResponse) IsAsync() bool {
	return r.Async
}

// GetOperationKey returns the operation key of the asynchronous request.
func (r *UpdateInstanceResponse) GetOperationKey() *OperationKey {
	return r.OperationKey
}

// GetDashboardURL returns the dashboard URL if provided by the broker.
func (r *UpdateInstanceResponse) GetDashboardURL() *string {
	return r.DashboardURL
}
