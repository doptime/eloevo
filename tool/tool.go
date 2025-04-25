package tool

import (
	"encoding/json"
	"log"
	"reflect"

	"github.com/doptime/eloevo/memory"
	openai "github.com/sashabaranov/go-openai"
)

type ToolInterface interface {
	HandleCallback(Param interface{}, CallMemory map[string]any) (err error)
	OaiTool() *openai.Tool
	Name() string
}

// Tool 是FuctionCall的逻辑实现。FunctionCall 是Tool的接口定义
type Tool[v any] struct {
	openai.Tool
	MemoryCacheKey string
	Functions      []func(param v)
}

func (t *Tool[v]) OaiTool() *openai.Tool {
	return &t.Tool
}
func (t *Tool[v]) Name() string {
	return t.Tool.Function.Name
}

func (t *Tool[v]) WithFunction(f func(param v)) *Tool[v] {
	t.Functions = append(t.Functions, f)
	return t
}

func (t *Tool[v]) WithMemoryCacheKey(key string) *Tool[v] {
	t.MemoryCacheKey = key
	return t
}
func (t *Tool[v]) HandleCallback(Param interface{}, CallMemory map[string]any) (err error) {
	var parambytes []byte
	if str, ok := Param.(string); ok {
		parambytes = []byte(str)
	} else {
		parambytes, err = json.Marshal(Param)
		if err != nil {
			log.Printf("Error parsing arguments for tool %s: %v", t.Tool.Function.Name, err)
		}
	}

	var val v
	vType := reflect.TypeOf(val) // Decode escaped Unicode in Param
	if vType.Kind() == reflect.Ptr {
		// Create a new instance of the pointed type
		valPtr := reflect.New(vType.Elem()).Interface()
		err = json.Unmarshal(parambytes, valPtr)
		if err != nil {
			log.Printf("Error parsing arguments for tool %s: %v", t.Tool.Function.Name, err)
			return err
		}
		memory.SaveToShareMemory(t.MemoryCacheKey, reflect.ValueOf(valPtr).Interface())
		// Assign the dereferenced pointer to val
		val = reflect.ValueOf(valPtr).Interface().(v)
		//set val field named Extra to CallMemory
		if CallMemory != nil {
			valType := reflect.TypeOf(val).Elem()
			valValue := reflect.ValueOf(val).Elem()
			for i := 0; i < valType.NumField(); i++ {
				field := valType.Field(i)
				if field.Name == "Extra" {
					valValue.Field(i).Set(reflect.ValueOf(CallMemory))
				}
			}
		}
	} else {
		// Unmarshal directly into val
		err = json.Unmarshal(parambytes, &val)
		if err != nil {
			log.Printf("Error parsing arguments for tool %s: %v", t.Tool.Function.Name, err)
			return err
		}
		memory.SaveToShareMemory(t.MemoryCacheKey, val)
		//set val field named Extra to CallMemory
		if CallMemory != nil {
			valType := reflect.TypeOf(val)
			valValue := reflect.ValueOf(val)
			for i := 0; i < valType.NumField(); i++ {
				field := valType.Field(i)
				if field.Name == "Extra" {
					valValue.Field(i).Set(reflect.ValueOf(CallMemory))
				}
			}
		}
	}

	for _, f := range t.Functions {
		f(val)
	}
	return nil
}

func NewTool[v any](name string, description string, fs ...func(param v)) *Tool[v] {
	// Inspect the type of v , should be a struct
	vType := reflect.TypeOf(new(v)).Elem()
	vValue := reflect.ValueOf(new(v)).Elem()

	for vType.Kind() == reflect.Ptr {
		vType = vType.Elem()
		vValue = vValue.Elem()
	}

	params := make(map[string]any)
	if vType.Kind() == reflect.Struct {
		// Map parameter fields to JSON schema definitions
		for i := 0; i < vType.NumField(); i++ {
			field := vType.Field(i)
			def := map[string]string{
				"type":        mapKindToDataType(field.Type.Kind()),
				"description": field.Tag.Get("description"),
			}
			if def["description"] == "-" || def["description"] == "" {
				continue
			}
			params[field.Name] = def
		}
	}

	a := &Tool[v]{
		Tool: openai.Tool{Type: openai.ToolTypeFunction, Function: &openai.FunctionDefinition{
			Name:        name,
			Description: description,
			Parameters:  params,
		}},
		Functions: fs,
	}

	// Define the function to handle LLM response
	//HandleFuncs[name] = a

	return a
}

func mapKindToDataType(kind reflect.Kind) string {
	var mapKindToDataType = map[reflect.Kind]string{
		reflect.Struct:  "object",
		reflect.Float32: "number", reflect.Float64: "number",
		reflect.Int: "integer", reflect.Int8: "integer", reflect.Int16: "integer", reflect.Int32: "integer", reflect.Int64: "integer",
		reflect.Uint: "integer", reflect.Uint8: "integer", reflect.Uint16: "integer", reflect.Uint32: "integer", reflect.Uint64: "integer",
		reflect.String:  "string",
		reflect.Slice:   "array",
		reflect.Array:   "array",
		reflect.Bool:    "boolean",
		reflect.Invalid: "null",
		reflect.Map:     "object",
	}
	return mapKindToDataType[kind]
}
