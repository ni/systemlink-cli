package commandline_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ni/systemlink-cli/internal/commandline"
	"github.com/ni/systemlink-cli/internal/model"
)

var convertValuesTest = []struct {
	value          string
	targetType     model.ParameterType
	convertedValue interface{}
	err            error
}{
	{"hello world", model.StringType, "hello world", nil},
	{"123", model.IntegerType, 123, nil},
	{"001", model.IntegerType, 1, nil},
	{"123.123", model.IntegerType, nil, fmt.Errorf("Invalid value for argument 'my-value'")},
	{"invalid", model.IntegerType, nil, fmt.Errorf("Invalid value for argument 'my-value'")},
	{"123", model.NumberType, 123.0, nil},
	{"001", model.NumberType, 1.0, nil},
	{"99.678", model.NumberType, 99.678, nil},
	{"invalid", model.NumberType, nil, fmt.Errorf("Invalid value for argument 'my-value'")},
	{"true", model.BooleanType, true, nil},
	{"false", model.BooleanType, false, nil},
	{"invalid", model.BooleanType, nil, fmt.Errorf("Invalid value for argument 'my-value'")},
	{"true", model.ObjectType, true, nil},
	{"1.1", model.ObjectType, 1.1, nil},
	{`{ "hello": "world" }`, model.ObjectType, map[string]interface{}{"hello": "world"}, nil},
	{`["hello", "world"]`, model.ObjectType, []interface{}{"hello", "world"}, nil},
	{"{invalid}", model.ObjectType, nil, fmt.Errorf("Invalid value for argument 'my-value'")},
	{"hello,world,123", model.StringArrayType, []string{"hello", "world", "123"}, nil},
	{"hello,world,", model.StringArrayType, []string{"hello", "world", ""}, nil},
	{"1,2,3", model.IntegerArrayType, []int{1, 2, 3}, nil},
	{"1,2,invalid", model.IntegerArrayType, nil, fmt.Errorf("Invalid value for argument 'my-value'")},
	{"1.1,2.2,3.3", model.NumberArrayType, []float64{1.1, 2.2, 3.3}, nil},
	{"1.1,2.2,invalid", model.IntegerArrayType, nil, fmt.Errorf("Invalid value for argument 'my-value'")},
	{"true,false,true", model.BooleanArrayType, []bool{true, false, true}, nil},
	{"trUE,TRUE,False", model.BooleanArrayType, []bool{true, true, false}, nil},
	{"true,invalid,false", model.BooleanArrayType, nil, fmt.Errorf("Invalid value for argument 'my-value'")},
	{`["hello", "world"]`, model.ObjectArrayType, []interface{}{"hello", "world"}, nil},
	{`["invalid"}`, model.ObjectType, nil, fmt.Errorf("Invalid value for argument 'my-value'")},
}

func TestConvertValues(t *testing.T) {
	for _, tt := range convertValuesTest {
		converter := commandline.ValueConverter{}

		values := map[string]string{
			"my-value": tt.value,
		}
		parameters := []model.Parameter{
			{Name: "my-value", TypeInfo: tt.targetType},
		}
		result, err := converter.ConvertValues(values, parameters)

		if tt.err != nil && !reflect.DeepEqual(tt.err, err) {
			t.Errorf("Error output was wrong, Expected: %s but got: %s", tt.err, err)
		}
		if tt.convertedValue != nil && !reflect.DeepEqual(tt.convertedValue, result[0].Value) {
			t.Errorf("Converted value was wrong, Expected: %s but got: %s", tt.convertedValue, result[0].Value)
		}
	}
}

func TestConvertMultipleValues(t *testing.T) {
	converter := commandline.ValueConverter{}

	values := map[string]string{
		"my-value-1": "true",
		"my-value-2": "1",
	}
	parameters := []model.Parameter{
		{Name: "my-value-1", TypeInfo: model.BooleanType},
		{Name: "my-value-2", TypeInfo: model.IntegerType},
	}
	result, _ := converter.ConvertValues(values, parameters)

	if len(result) != 2 {
		t.Errorf("Expected 2 converted values but got: %v", len(result))
	}
	for _, r := range result {
		if r.Name != "my-value-1" && r.Name != "my-value-2" {
			t.Errorf("Return parameter name was wrong, got: %v", r.Name)
		}
		if r.Name == "my-value-1" && r.Value != true {
			t.Errorf("Converted value was wrong, Expected: %v but got: %v", true, r.Value)
		}
		if r.Name == "my-value-2" && r.Value != 1 {
			t.Errorf("Converted value was wrong, Expected: %v but got: %v", 1, r.Value)
		}
	}
}

func TestConvertMultipleParametersWithSameName(t *testing.T) {
	converter := commandline.ValueConverter{}

	values := map[string]string{
		"my-value": "foo",
	}
	parameters := []model.Parameter{
		{Name: "my-value", TypeInfo: model.StringType, Location: model.QueryLocation},
		{Name: "my-value", TypeInfo: model.StringType, Location: model.PathLocation},
	}
	result, _ := converter.ConvertValues(values, parameters)

	if len(result) != 2 {
		t.Errorf("Expected 2 converted values but got: %v", len(result))
	}

	if result[0].Location != model.QueryLocation && result[0].Location != model.PathLocation {
		t.Errorf("Return parameter location was wrong, got: %v", result[0].Location)
	}
	if result[1].Location != model.QueryLocation && result[1].Location != model.PathLocation {
		t.Errorf("Return parameter location was wrong, got: %v", result[1].Location)
	}
	if result[0].Location == result[1].Location {
		t.Errorf("Return parameter location should be different, got: %v for both", result[0].Location)
	}
}

func TestConvertUndefinedParameterThrows(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The converter did not throw")
		}
	}()

	converter := commandline.ValueConverter{}

	values := map[string]string{
		"my-value-1": "true",
	}
	parameters := []model.Parameter{}
	converter.ConvertValues(values, parameters)
}
