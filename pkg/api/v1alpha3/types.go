package v1alpha3

// CopyImageSchema
type CopyImageSchema struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

// Response schema
type Response struct {
	Name       string           `json:"name"`
	StatusCode string           `json:"statuscode"`
	Status     string           `json:"status"`
	Message    string           `json:"message"`
	Payload    *SchemaInterface `json:"payload,omitempty"`
}

// GenericSchema - used in the GenericHandler (complex data object)
type GenericSchema struct {
	Url string
}

// All the go microservices will using this schema
type SchemaInterface struct {
	ID         int64  `json:"_id,omitempty"`
	LastUpdate int64  `json:"lastupdate,omitempty"`
	MetaInfo   string `json:"metainfo,omitempty"`
}
