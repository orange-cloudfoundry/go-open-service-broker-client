package v2

// IsAsync returns true if the provision request is being handled asynchronously.
func (r *ProvisionResponse) IsAsync() bool {
	return r.Async
}

// GetOperationKey returns the operation key for the asynchronous provision request.
func (r *ProvisionResponse) GetOperationKey() *OperationKey {
	return r.OperationKey
}

// GetDashboardURL returns the dashboard URL provided by the broker if available.
func (r *ProvisionResponse) GetDashboardURL() *string {
	return r.DashboardURL
}
