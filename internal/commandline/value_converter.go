package commandline

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ni/systemlink-cli/internal/model"
)

// ValueConverter provides functions to convert between input data
// of the command line and the data types in the model
type ValueConverter struct{}

func (c ValueConverter) convertToIntegerArray(value string) ([]int, error) {
	var result []int
	for _, v := range strings.Split(value, ",") {
		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		result = append(result, i)
	}
	return result, nil
}

func (c ValueConverter) convertToNumberArray(value string) ([]float64, error) {
	var result []float64
	for _, v := range strings.Split(value, ",") {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, err
		}
		result = append(result, f)
	}
	return result, nil
}

func (c ValueConverter) convertToBooleanArray(value string) ([]bool, error) {
	var result []bool
	for _, v := range strings.Split(value, ",") {
		b, err := c.convertToBoolean(v)
		if err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, nil
}

func (c ValueConverter) convertToBoolean(value string) (bool, error) {
	if strings.EqualFold(value, "true") {
		return true, nil
	} else if strings.EqualFold(value, "false") {
		return false, nil
	}
	return false, fmt.Errorf("Cannot convert %s to boolean", value)
}

func (c ValueConverter) convertToType(value string, typeInfo model.ParameterType) (interface{}, error) {
	switch typeInfo {
	case model.BooleanType:
		return c.convertToBoolean(value)
	case model.IntegerType:
		return strconv.Atoi(value)
	case model.NumberType:
		return strconv.ParseFloat(value, 64)
	case model.StringArrayType:
		return strings.Split(value, ","), nil
	case model.IntegerArrayType:
		return c.convertToIntegerArray(value)
	case model.NumberArrayType:
		return c.convertToNumberArray(value)
	case model.BooleanArrayType:
		return c.convertToBooleanArray(value)
	case model.ObjectType, model.ObjectArrayType:
		var j interface{}
		err := json.Unmarshal([]byte(value), &j)
		return j, err
	}
	return value, nil
}

func (c ValueConverter) findParameter(name string, parameters []model.Parameter) model.Parameter {
	for _, p := range parameters {
		if p.Name == name {
			return p
		}
	}
	panic(fmt.Sprintf("Parameter %s not defined in model.", name))
}

func (c ValueConverter) findParameters(name string, parameters []model.Parameter) []model.Parameter {
	var result []model.Parameter
	for _, p := range parameters {
		if p.Name == name {
			result = append(result, p)
		}
	}

	if len(result) == 0 {
		panic(fmt.Sprintf("Parameter %s not defined in model.", name))
	}
	return result
}

func (c ValueConverter) convertValue(value string, parameters []model.Parameter) ([]model.ParameterValue, error) {
	var result []model.ParameterValue
	for _, param := range parameters {
		convertedValue, convertErr := c.convertToType(value, param.TypeInfo)
		if convertErr != nil {
			return nil, fmt.Errorf("Invalid value for argument '%s'", param.Name)
		}
		parameterValue := model.ParameterValue{Parameter: param, Value: convertedValue}
		result = append(result, parameterValue)
	}
	return result, nil
}

// ConvertValues converts the given input parameter strings into to defined types
// of the model parameters
func (c ValueConverter) ConvertValues(values map[string]string, parameters []model.Parameter) ([]model.ParameterValue, error) {
	var result []model.ParameterValue

	for key, value := range values {
		params := c.findParameters(key, parameters)
		convertedValues, err := c.convertValue(value, params)
		if err != nil {
			return nil, err
		}
		result = append(result, convertedValues...)
	}

	return result, nil
}
