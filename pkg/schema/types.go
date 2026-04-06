package schema

// TypeMappings maps Go types to OpenAPI types
var TypeMappings = map[string]Schema{
	"string":  {Type: "string"},
	"int":     {Type: "integer"},
	"int8":    {Type: "integer"},
	"int16":   {Type: "integer"},
	"int32":   {Type: "integer", Format: "int32"},
	"int64":   {Type: "integer", Format: "int64"},
	"uint":    {Type: "integer"},
	"uint8":   {Type: "integer"},
	"uint16":  {Type: "integer"},
	"uint32":  {Type: "integer", Format: "int32"},
	"uint64":  {Type: "integer", Format: "int64"},
	"float32": {Type: "number", Format: "float"},
	"float64": {Type: "number", Format: "double"},
	"bool":    {Type: "boolean"},
	"byte":    {Type: "string", Format: "byte"},
}

// FormatMappings maps common formats
var FormatMappings = map[string]string{
	"uuid":      "uuid",
	"email":     "email",
	"uri":       "uri",
	"url":       "uri",
	"hostname":  "hostname",
	"ipv4":      "ipv4",
	"ipv6":      "ipv6",
	"date":      "date",
	"date-time": "date-time",
	"time":      "time",
	"password":  "password",
	"byte":      "byte",
	"binary":    "binary",
}

// GetTypeMapping returns the OpenAPI type for a Go type
func GetTypeMapping(goType string) *Schema {
	if schema, ok := TypeMappings[goType]; ok {
		return &schema
	}
	return &Schema{Type: "object"}
}
