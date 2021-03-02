package vo

type ServiceID string

type Service struct {
	ID             ServiceID
	BackendAddress string
	Path           string
	MimeTypes      []string
}

type ServiceError struct {
	Err string
}

func (e *ServiceError) Error() string {
	return e.Err
}

type ServerURL string

type PathOrMime string

type ClientConfig []*Service
type MultiServerClientConfig map[ServerURL]ClientConfig
