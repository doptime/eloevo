package tool

import (
	"encoding/json"
	"log"
	"reflect"

	"github.com/mitchellh/mapstructure"
	openai "github.com/sashabaranov/go-openai"

	genai "google.golang.org/genai"
)

type ToolInterface interface {
	HandleCallback(Param interface{}, CallMemory map[string]any) (err error)
	OaiTool() *openai.Tool
	GoogleGenaiTool() *genai.FunctionDeclaration
	Name() string
}

// Tool 是FuctionCall的逻辑实现。FunctionCall 是Tool的接口定义
type Tool[v any] struct {
	openai.Tool
	GoogleFunc genai.FunctionDeclaration
	Functions  []func(param v)
}

func (t *Tool[v]) OaiTool() *openai.Tool {
	return &t.Tool
}
func (t *Tool[v]) GoogleGenaiTool() *genai.FunctionDeclaration {
	return &t.GoogleFunc
}
func (t *Tool[v]) Name() string {
	return t.Tool.Function.Name
}

func (t *Tool[v]) WithFunction(f func(param v)) *Tool[v] {
	t.Functions = append(t.Functions, f)
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
	err = json.Unmarshal(parambytes, &val) // 直接反序列化到 v 的地址
	if err != nil {
		log.Printf("Error parsing arguments for tool %s: %v。make sure type of v is a ponter to struct", t.Tool.Function.Name, err)
		return err
	}
	//Extract the memory cached key to destination struct
	if CallMemory != nil {
		mapstructure.Decode(CallMemory, val)
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
	var goooglefunctionDeclarations genai.FunctionDeclaration = genai.FunctionDeclaration{
		Name:        name,
		Description: description,
		Parameters:  &genai.Schema{},
	}
	if vType.Kind() == reflect.Struct {
		// Map parameter fields to JSON schema definitions
		//build openai.Tool
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
		//build google genai.Tool
		goooglefunctionDeclarations.Parameters.Type = genai.TypeObject
		for i := 0; i < vType.NumField(); i++ {
			field := vType.Field(i)

			if goooglefunctionDeclarations.Description == "-" || goooglefunctionDeclarations.Description == "" {
				continue
			}
			var Parameters genai.Schema
			Parameters.Title = field.Name
			Parameters.Type = genai.TypeUnspecified
			if field.Type.Kind() == reflect.Struct {
				Parameters.Type = genai.TypeObject
			} else if field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Array {
				Parameters.Type = genai.TypeArray
			} else if field.Type.Kind() == reflect.Map {
				Parameters.Type = genai.TypeObject
			} else if field.Type.Kind() == reflect.String {
				Parameters.Type = genai.TypeString
			} else if field.Type.Kind() == reflect.Int || field.Type.Kind() == reflect.Int8 || field.Type.Kind() == reflect.Int16 || field.Type.Kind() == reflect.Int32 || field.Type.Kind() == reflect.Int64 ||
				field.Type.Kind() == reflect.Uint || field.Type.Kind() == reflect.Uint8 || field.Type.Kind() == reflect.Uint16 || field.Type.Kind() == reflect.Uint32 || field.Type.Kind() == reflect.Uint64 {
				Parameters.Type = genai.TypeInteger
			} else if field.Type.Kind() == reflect.Float32 || field.Type.Kind() == reflect.Float64 {
				Parameters.Type = genai.TypeNumber
			} else if field.Type.Kind() == reflect.Bool {
				Parameters.Type = genai.TypeBoolean
			} else if field.Type.Kind() == reflect.Invalid {
				Parameters.Type = genai.TypeNULL
			}
			Parameters.Description = field.Tag.Get("description")

			goooglefunctionDeclarations.Parameters.Properties[field.Name] = &Parameters
		}
	}

	a := &Tool[v]{
		Tool: openai.Tool{Type: openai.ToolTypeFunction, Function: &openai.FunctionDefinition{
			Name:        name,
			Description: description,
			Parameters:  params,
		}},
		GoogleFunc: goooglefunctionDeclarations,
		Functions:  fs,
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
