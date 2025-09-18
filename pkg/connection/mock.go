package connection

// MockConnection is a simple mock implementation for testing
type MockConnection struct {
	Name string
}

func (m *MockConnection) Authenticate() error {
	return nil
}

func (m *MockConnection) FetchReport(params map[string]string) ([]map[string]string, error) {
	// Return some mock data
	return []map[string]string{
		{"id": "1", "name": "Test User 1"},
		{"id": "2", "name": "Test User 2"},
	}, nil
}
