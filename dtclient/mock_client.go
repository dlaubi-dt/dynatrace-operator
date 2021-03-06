package dtclient

import "github.com/stretchr/testify/mock"

// MockDynatraceClient implements a Dynatrace REST API Client mock
type MockDynatraceClient struct {
	mock.Mock
}

func (o *MockDynatraceClient) GetTenantInfo() (*TenantInfo, error) {
	args := o.Called()
	return args.Get(0).(*TenantInfo), args.Error(1)
}

func (o *MockDynatraceClient) QueryOutdatedActiveGates(query *ActiveGateQuery) ([]ActiveGate, error) {
	args := o.Called(query)
	return args.Get(0).([]ActiveGate), args.Error(1)
}

func (o *MockDynatraceClient) QueryActiveGates(query *ActiveGateQuery) ([]ActiveGate, error) {
	args := o.Called(query)
	return args.Get(0).([]ActiveGate), args.Error(1)
}

func (o *MockDynatraceClient) GetAgentVersionForIP(ip string) (string, error) {
	args := o.Called(ip)
	return args.String(0), args.Error(1)
}

func (o *MockDynatraceClient) GetLatestAgentVersion(os, installerType string) (string, error) {
	args := o.Called(os, installerType)
	return args.String(0), args.Error(1)
}

func (o *MockDynatraceClient) GetConnectionInfo() (ConnectionInfo, error) {
	args := o.Called()
	return args.Get(0).(ConnectionInfo), args.Error(1)
}

func (o *MockDynatraceClient) GetCommunicationHostForClient() (CommunicationHost, error) {
	args := o.Called()
	return args.Get(0).(CommunicationHost), args.Error(1)
}

func (o *MockDynatraceClient) SendEvent(event *EventData) error {
	args := o.Called(event)
	return args.Error(0)
}

func (o *MockDynatraceClient) GetEntityIDForIP(ip string) (string, error) {
	args := o.Called(ip)
	return args.String(0), args.Error(1)
}

func (o *MockDynatraceClient) GetTokenScopes(token string) (TokenScopes, error) {
	args := o.Called(token)
	return args.Get(0).(TokenScopes), args.Error(1)
}

func (o *MockDynatraceClient) GetClusterInfo() (*ClusterInfo, error) {
	args := o.Called()
	return args.Get(0).(*ClusterInfo), args.Error(1)
}

func (o *MockDynatraceClient) AddToDashboard(label string, kubernetesApiEndpoint string, bearerToken string) (string, error) {
	args := o.Called(label, kubernetesApiEndpoint, bearerToken)
	return args.Get(0).(string), args.Error(1)
}
