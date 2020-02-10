package commandline

import "github.com/ni/systemlink-cli/internal/model"

// ServiceCaller interface  abstracts calling the external services.
// The Call function takes in a model describing the API of the service
// as well as the provided parameters.
type ServiceCaller interface {
	Call(operation model.Operation, parameterValues []model.ParameterValue, settings model.Settings) (int, string, error)
}
