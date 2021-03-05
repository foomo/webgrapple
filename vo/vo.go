package vo

// ServiceID an identifier for a service
type ServiceID string

// Service a service to proxy to
type Service struct {
	ID      ServiceID
	Address string
	Custom  map[string]interface{}
}

// ServiceError an error used in client server communication
type ServiceError struct {
	Err string
}

func (e *ServiceError) Error() string {
	return e.Err
}

// ClientConfig configuration for a client to use for webgrapple reverse proxy
type ClientConfig []*Service
