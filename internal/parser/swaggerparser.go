package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"

	"github.com/ni/systemlink-cli/internal/model"
)

const defaultSystemLinkURL = "https://api.systemlinkcloud.com"

// SwaggerParser implements the Parser interface and turns a swagger yaml file
// into the internal Definition structure which describes all operations
// and parameters of the service
type SwaggerParser struct{}

func (p SwaggerParser) parseURL(spec *spec.Swagger) string {
	var url = defaultSystemLinkURL
	if spec.Host != "" && len(spec.Schemes) > 0 {
		url = spec.Schemes[0] + "://" + spec.Host
	}
	return url
}

func (p SwaggerParser) contains(value string, values []string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func (p SwaggerParser) parseArrayType(typeInfo string) (model.ParameterType, error) {
	switch typeInfo {
	case "string":
		return model.StringArrayType, nil
	case "integer":
		return model.IntegerArrayType, nil
	case "number":
		return model.NumberArrayType, nil
	case "boolean":
		return model.BooleanArrayType, nil
	case "object":
		return model.ObjectArrayType, nil
	}
	return 0, fmt.Errorf("Invalid array type '%s'", typeInfo)
}

func (p SwaggerParser) parseType(typeInfo string, items *spec.SchemaOrArray) (model.ParameterType, error) {
	switch typeInfo {
	case "string":
		return model.StringType, nil
	case "integer":
		return model.IntegerType, nil
	case "number":
		return model.NumberType, nil
	case "boolean":
		return model.BooleanType, nil
	case "object":
		return model.ObjectType, nil
	case "array":
		arrayType := "object"
		if items != nil && items.Schema != nil && len(items.Schema.Type) > 0 {
			arrayType = items.Schema.Type[0]
		}
		return p.parseArrayType(arrayType)
	}
	return 0, fmt.Errorf("Invalid type '%s'", typeInfo)
}

func (p SwaggerParser) parseLocation(in string) (model.ParameterLocation, error) {
	switch in {
	case "body":
		return model.BodyLocation, nil
	case "path":
		return model.PathLocation, nil
	case "query":
		return model.QueryLocation, nil
	case "header":
		return model.HeaderLocation, nil
	}
	return 0, fmt.Errorf("Invalid location '%s'", in)
}

func (p SwaggerParser) parseParameter(param spec.Parameter) (*model.Parameter, error) {
	name := param.Name
	description := param.Description
	required := param.Required
	typeInfo, err := p.parseType(param.Type, nil)
	if err != nil {
		return nil, err
	}
	location, err := p.parseLocation(param.In)
	if err != nil {
		return nil, err
	}

	return &model.Parameter{
		Name:        name,
		Description: description,
		TypeInfo:    typeInfo,
		Location:    location,
		Required:    required,
	}, nil
}

func (p SwaggerParser) parseProperties(schema *spec.Schema, location model.ParameterLocation) ([]model.Parameter, error) {
	var result []model.Parameter

	for name, property := range schema.Properties {
		propertyType := "object"
		if len(property.Type) > 0 {
			propertyType = property.Type[0]
		}
		typeInfo, err := p.parseType(propertyType, property.Items)
		if err != nil {
			return nil, err
		}
		description := property.Description
		required := p.contains(name, schema.Required)

		var param = model.Parameter{
			Name:        name,
			Description: description,
			TypeInfo:    typeInfo,
			Location:    location,
			Required:    required,
		}
		result = append(result, param)
	}

	return result, nil
}

func (p SwaggerParser) parseArraysAndProperties(param spec.Parameter) ([]model.Parameter, error) {
	var result []model.Parameter

	schema := param.Schema
	location, err := p.parseLocation(param.In)
	if err != nil {
		return nil, err
	}

	properties, err := p.parseProperties(schema, location)
	if err != nil {
		return nil, err
	}
	result = append(result, properties...)

	if schema.Items != nil {
		var arrayItemSchema = schema.Items.Schema
		if arrayItemSchema.Type[0] != "object" {
			typeInfo, err := p.parseType(schema.Type[0], schema.Items)
			if err != nil {
				return nil, err
			}

			param := model.Parameter{
				Name:        param.Name,
				Description: param.Description,
				TypeInfo:    typeInfo,
				Location:    location,
				Required:    param.Required,
			}
			result = append(result, param)
		}
		properties, err := p.parseProperties(arrayItemSchema, location)
		if err != nil {
			return nil, err
		}
		result = append(result, properties...)
	}

	return result, nil
}

func (p SwaggerParser) parseParameters(params []spec.Parameter) ([]model.Parameter, error) {
	var result []model.Parameter

	for _, param := range params {
		if param.Schema != nil {
			parameters, err := p.parseArraysAndProperties(param)
			if err != nil {
				return nil, err
			}
			result = append(result, parameters...)
		} else {
			parameter, err := p.parseParameter(param)
			if err != nil {
				return nil, err
			}
			result = append(result, *parameter)
		}
	}

	return result, nil
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func (p SwaggerParser) toDashCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}-${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}-${2}")
	return strings.ToLower(snake)
}

func (p SwaggerParser) parseMethodName(operation string, path string) string {
	if operation == "" {
		operation = strings.Replace(path, "/", "", -1)
	}
	var splitted = strings.Split(operation, ".")
	var name = strings.Replace(splitted[len(splitted)-1], "_", "-", -1)
	return p.toDashCase(name)
}

func (p SwaggerParser) caseInsensitiveContains(a, b string) bool {
	return strings.Contains(strings.ToUpper(a), strings.ToUpper(b))
}

func (p SwaggerParser) parseOperation(method string, path string, operation *spec.Operation) (*model.Operation, error) {
	if operation == nil {
		return nil, nil
	}
	if p.caseInsensitiveContains(path, "websocket") {
		return nil, nil
	}

	name := p.parseMethodName(operation.ID, path)
	description := operation.Description
	parameters, err := p.parseParameters(operation.Parameters)
	if err != nil {
		return nil, err
	}

	return &model.Operation{
		Name:        name,
		Description: description,
		Parameters:  parameters,
		Method:      method,
		Path:        path,
	}, nil
}

func (p SwaggerParser) parseOperations(path string, pathItem spec.PathItem) ([]model.Operation, error) {
	var result []model.Operation
	methods := []struct {
		method    string
		operation *spec.Operation
	}{
		{"GET", pathItem.Get},
		{"PUT", pathItem.Put},
		{"POST", pathItem.Post},
		{"DELETE", pathItem.Delete},
		{"OPTIONS", pathItem.Options},
		{"HEAD", pathItem.Head},
		{"PATCH", pathItem.Patch},
	}

	for _, m := range methods {
		operation, err := p.parseOperation(m.method, path, m.operation)
		if err != nil {
			return nil, err
		}

		if operation != nil {
			result = append(result, *operation)
		}
	}

	return result, nil
}

func (p SwaggerParser) parsePaths(basePath string, paths *spec.Paths) ([]model.Operation, error) {
	var result []model.Operation

	if paths != nil {
		for path, pathItem := range paths.Paths {
			ops, err := p.parseOperations(basePath+path, pathItem)
			if err != nil {
				return nil, err
			}
			result = append(result, ops...)
		}
	}

	return result, nil
}

func (p SwaggerParser) parse(m model.Data) (*model.Definition, error) {
	document, err := loads.Analyzed(m.Content, "2.0")
	if err != nil {
		return nil, NewParseError(m.Name, err)
	}
	document, err = document.Expanded()
	if err != nil {
		return nil, NewParseError(m.Name, err)
	}
	spec := document.Spec()
	url := p.parseURL(spec)
	operations, err := p.parsePaths(spec.BasePath, spec.Paths)
	if err != nil {
		return nil, NewParseError(m.Name, err)
	}
	return &model.Definition{Name: m.Name, URL: url, Operations: operations}, nil
}

// Parse takes a list of model byte streams which need to contain valid swagger yaml
// and turns it into a list of Definition's
func (p SwaggerParser) Parse(models []model.Data) ([]model.Definition, error) {
	var definitions = make([]model.Definition, len(models))
	for i, m := range models {
		definition, err := p.parse(m)
		if err != nil {
			return nil, err
		}
		definitions[i] = *definition
	}
	return definitions, nil
}
