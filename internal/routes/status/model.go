package status

type DatabaseInfo struct {
	Version           string `json:"version"`
	MaxConnections    int16  `json:"max_connections"`
	OpenedConnections int16  `json:"opened_connections"`
}

type StatusInfo struct {
	UpdatedAt    string       `json:"updated_at"`
	Dependencies Dependencies `json:"dependencies"`
}

type Dependencies struct {
	Database DatabaseInfo `json:"database"`
}
