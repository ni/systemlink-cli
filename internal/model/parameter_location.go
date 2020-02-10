package model

// ParameterLocation describes where a parameter value is located
type ParameterLocation int

const (
	// PathLocation means the parameter is stored in the path of the URL
	// e.g. /nitag/v1/tags/<path>
	PathLocation ParameterLocation = 1 + iota
	// BodyLocation means the parameter is stored in the message body
	// e.g. { path: "<path>", type: "INT" }
	BodyLocation
	// QueryLocation means the parameter is stored in the query string
	// e.g. /nitag/v1/tags?path=<path>
	QueryLocation
	// HeaderLocation means the parameter is stored in the http header
	// e.g. x-ni-api-key: <api-key>
	HeaderLocation
)
