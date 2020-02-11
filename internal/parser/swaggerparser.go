package parser

import (
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

func (p SwaggerParser) parseArrayType(typeInfo string) model.ParameterType {
	switch typeInfo {
	case "string":
		return model.StringArrayType
	case "integer":
		return model.IntegerArrayType
	case "number":
		return model.NumberArrayType
	case "boolean":
		return model.BooleanArrayType
	case "object":
		return model.ObjectArrayType
	}
	panic("Unknown array type found in swagger model: " + typeInfo)
}

func (p SwaggerParser) parseType(typeInfo string, items *spec.SchemaOrArray) model.ParameterType {
	switch typeInfo {
	case "string":
		return model.StringType
	case "integer":
		return model.IntegerType
	case "number":
		return model.NumberType
	case "boolean":
		return model.BooleanType
	case "object":
		return model.ObjectType
	case "array":
		arrayType := "object"
		if len(items.Schema.Type) > 0 {
			arrayType = items.Schema.Type[0]
		}
		return p.parseArrayType(arrayType)
	}
	panic("Unknown type found in swagger model: " + typeInfo)
}

func (p SwaggerParser) parseLocation(in string) model.ParameterLocation {
	switch in {
	case "body":
		return model.BodyLocation
	case "path":
		return model.PathLocation
	case "query":
		return model.QueryLocation
	case "header":
		return model.HeaderLocation
	}
	panic("Unknown location found in swagger model: " + in)
}

func (p SwaggerParser) parseParameter(param spec.Parameter) model.Parameter {
	var name = param.Name
	var description = param.Description
	var required = param.Required
	var typeInfo = p.parseType(param.Type, nil)
	var location = p.parseLocation(param.In)

	return model.Parameter{
		Name:        name,
		Description: description,
		TypeInfo:    typeInfo,
		Location:    location,
		Required:    required,
	}
}

func (p SwaggerParser) parseProperties(schema *spec.Schema, location model.ParameterLocation) []model.Parameter {
	var result []model.Parameter

	for name, property := range schema.Properties {
		propertyType := "object"
		if len(property.Type) > 0 {
			propertyType = property.Type[0]
		}
		var typeInfo = p.parseType(propertyType, property.Items)
		var description = property.Description
		var required = p.contains(name, schema.Required)

		var p = model.Parameter{
			Name:        name,
			Description: description,
			TypeInfo:    typeInfo,
			Location:    location,
			Required:    required,
		}
		result = append(result, p)
	}

	return result
}

func (p SwaggerParser) parseArraysAndProperties(param spec.Parameter) []model.Parameter {
	var result []model.Parameter

	schema := param.Schema
	location := p.parseLocation(param.In)

	var properties = p.parseProperties(schema, location)
	result = append(result, properties...)

	if schema.Items != nil {
		var arrayItemSchema = schema.Items.Schema
		if arrayItemSchema.Type[0] != "object" {
			var typeInfo = p.parseType(schema.Type[0], schema.Items)

			var p = model.Parameter{
				Name:        param.Name,
				Description: param.Description,
				TypeInfo:    typeInfo,
				Location:    location,
				Required:    param.Required,
			}
			result = append(result, p)
		}
		var properties = p.parseProperties(arrayItemSchema, location)
		result = append(result, properties...)
	}

	return result
}

func (p SwaggerParser) parseParameters(params []spec.Parameter) []model.Parameter {
	var result []model.Parameter

	for _, param := range params {
		if param.Schema != nil {
			var parameters = p.parseArraysAndProperties(param)
			result = append(result, parameters...)
		} else {
			var parameter = p.parseParameter(param)
			result = append(result, parameter)
		}
	}

	return result
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

func (p SwaggerParser) parseOperation(method string, path string, operation *spec.Operation) *model.Operation {
	if operation == nil {
		return nil
	}
	if p.caseInsensitiveContains(path, "websocket") {
		return nil
	}

	var name = p.parseMethodName(operation.ID, path)
	var description = operation.Description
	var parameters = p.parseParameters(operation.Parameters)

	return &model.Operation{
		Name:        name,
		Description: description,
		Parameters:  parameters,
		Method:      method,
		Path:        path,
	}
}

func (p SwaggerParser) parseOperations(path string, pathItem spec.PathItem) []model.Operation {
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
		var operation = p.parseOperation(m.method, path, m.operation)
		if operation != nil {
			result = append(result, *operation)
		}
	}

	return result
}

func (p SwaggerParser) parsePaths(basePath string, paths *spec.Paths) []model.Operation {
	var result []model.Operation

	if paths != nil {
		for path, pathItem := range paths.Paths {
			var ops = p.parseOperations(basePath+path, pathItem)
			result = append(result, ops...)
		}
	}

	return result
}

func (p SwaggerParser) parse(m model.Data) model.Definition {
	document, err := loads.Analyzed(m.Content, "2.0")
	if err != nil {
		panic(err)
	}
	document, err = document.Expanded()
	if err != nil {
		panic(err)
	}
	spec := document.Spec()
	var url = p.parseURL(spec)
	var operations = p.parsePaths(spec.BasePath, spec.Paths)
	return model.Definition{Name: m.Name, URL: url, Operations: operations}
}

// Parse takes a list of model byte streams which need to contain valid swagger yaml
// and turns it into a list of Definition's
func (p SwaggerParser) Parse(models []model.Data) []model.Definition {
	var definitions = make([]model.Definition, len(models))
	for i, m := range models {
		var definition = p.parse(m)
		definitions[i] = definition
	}
	return definitions
}
