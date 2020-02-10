package model

// ParameterValue contains the parameter model definition and its
// corresponding value
type ParameterValue struct {
	Parameter
	Value interface{}
}
