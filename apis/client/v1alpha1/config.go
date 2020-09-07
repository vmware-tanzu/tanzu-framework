package v1alpha1

// IsGlobal tells if the server is global.
func (s Server) IsGlobal() bool {
	if s.Type == GlobalServerType {
		return true
	}
	return false
}

// IsManagementCluster tells if the server is a management cluster.
func (s Server) IsManagementCluster() bool {
	if s.Type == ManagementClusterServerType {
		return true
	}
	return false
}
