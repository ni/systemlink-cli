package model

// Operation describes a single api and all its input parameters
// e.g. GET /nitag/v1/tags/{name}
//   - Method is GET
//   - Path is /nitag/v1/tags/{name}
//   - Parameters contains name as one of the input parameters
type Operation struct {
	Name        string
	Description string
	Parameters  []Parameter
	Method      string
	Path        string
}
