package commandline

import "github.com/ni/systemlink-cli/internal/model"

// Parser interface which turns the given byte stream into
// a structured service definition model with all declared operations.
type Parser interface {
	Parse(models []model.Data) ([]model.Definition, error)
}
