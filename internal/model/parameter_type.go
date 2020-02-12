package model

// ParameterType describes the data type of the parameter
type ParameterType int

const (
	// StringType means the parameter is a simple UTF8 string
	StringType ParameterType = 1 + iota
	// IntegerType means the parameter is a 32-bit integer
	IntegerType
	// NumberType means the parameter is a floating point number
	NumberType
	// BooleanType means the parameter is boolean
	BooleanType
	// ObjectType means the parameter is a generic object which is simply serialized
	ObjectType
	// FileType means the parameter is a file blob
	FileType
	// StringArrayType means the parameter is an string array
	StringArrayType
	// IntegerArrayType means the parameter is an integer array
	IntegerArrayType
	// NumberArrayType means the parameter is an number array
	NumberArrayType
	// BooleanArrayType means the parameter is an boolean array
	BooleanArrayType
	// ObjectArrayType means the parameter is an object array
	ObjectArrayType
)
