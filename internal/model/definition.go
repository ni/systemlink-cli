package model

// Definition contains the service URL and
// describes all operations the service provides.
type Definition struct {
	Name       string
	URL        string
	Operations []Operation
}
