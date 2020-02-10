package model

// Parameter contains the parsed metadata information of input parameters
type Parameter struct {
	Name        string
	Description string
	TypeInfo    ParameterType
	Location    ParameterLocation
	Required    bool
}
