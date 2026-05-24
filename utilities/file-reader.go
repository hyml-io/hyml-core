package utilities

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hyml-core/entities"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type FileReader struct {
}

var stringHandler = StringHandler{}

func (fileReader FileReader) ReadAllYamls(path string) []*entities.HymlDocument {

	yamlsArray := make([]*entities.HymlDocument, 0)

	yaml := fileReader.ReadYaml(path)

	if len(yaml.Def) > 0 {
		parsedTemplates, err := ParseTemplates(yaml)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Parsed templates:\n%#v\n", parsedTemplates)

		parsedVars, err := ParseVars(yaml)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Parsed Variables:\n%#v\n", parsedVars)

		parsedJsons, err := ParseJsons(yaml)

		var prettyJsons [][]byte

		for _, item := range parsedJsons {
			// Since item.Content is already a plain []byte, we can check its length
			if len(item.Content) == 0 {
				continue
			}

			var prettyBuf bytes.Buffer

			// Pass item.Content directly—no asterisks, no type casting needed
			err := json.Indent(&prettyBuf, item.Content, "", "  ")
			if err != nil {
				log.Fatalf("Invalid JSON syntax in file %s: %v", item.Path, err)
			}

			prettyJsons = append(prettyJsons, prettyBuf.Bytes())
		}

		fmt.Printf("Parsed JSONs:\n")

		for _, fileBytes := range prettyJsons {
			fmt.Print("\n" + string(fileBytes))
		}

	}

	yamlsArray = append(yamlsArray, yaml)

	return yamlsArray
}

func ParseVars(yaml *entities.HymlDocument) (map[string]entities.Var, error) {
	variables := make(map[string]entities.Var)

	extractedRawVars := extractVariables(yaml)

	for i, item := range extractedRawVars {
		// 1. Extract Name
		varNameVal, exists := item["var"]
		if !exists {
			return nil, fmt.Errorf("item at index %d is missing the required 'var' key", i)
		}
		name, ok := varNameVal.(string)
		if !ok {
			return nil, fmt.Errorf("item at index %d has a non-string 'var' key", i)
		}

		// 2. Extract Raw Value
		rawValue, hasValue := item["value"]
		if !hasValue {
			return nil, fmt.Errorf("variable '%s' is missing its 'value' key", name)
		}

		// 3. Extract Type string
		varType := ""
		if typeVal, hasType := item["type"]; hasType {
			if tStr, ok := typeVal.(string); ok {
				varType = tStr
			}
		}

		// 4. Parse the raw value into the target type
		parsedValue, err := castValueToType(rawValue, varType)
		if err != nil {
			return nil, fmt.Errorf("variable '%s' type mismatch: %w", name, err)
		}

		// 5. Determine Locked status
		locked := false
		if overwriteVal, hasOverwrite := item["enable-overwrite"]; hasOverwrite {
			if overwriteBool, isBool := overwriteVal.(bool); isBool && !overwriteBool {
				locked = true
			}
		}

		variables[name] = entities.Var{
			Name:   name,
			Type:   varType,
			Value:  parsedValue,
			Locked: locked,
		}
	}

	return variables, nil
}

func castValueToType(val any, targetType string) (any, error) {
	switch targetType {
	case "string":
		// Handle conversion if it came in as something else, or assert string
		return fmt.Sprintf("%v", val), nil

	case "number", "int", "integer":
		switch v := val.(type) {
		case int:
			return v, nil
		case float64:
			// Parsers like encoding/json decode all numbers as float64 by default
			return int(v), nil
		case string:
			// Fallback if numbers are quoted in the source file
			if i, err := strconv.Atoi(v); err == nil {
				return i, nil
			}
		}
		return nil, fmt.Errorf("cannot convert %v (%T) to integer", val, val)

	case "float", "double":
		switch v := val.(type) {
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f, nil
			}
		}
		return nil, fmt.Errorf("cannot convert %v (%T) to float64", val, val)

	case "bool", "boolean":
		if b, ok := val.(bool); ok {
			return b, nil
		}
		if s, ok := val.(string); ok {
			if b, err := strconv.ParseBool(s); err == nil {
				return b, nil
			}
		}
		return nil, fmt.Errorf("cannot convert %v (%T) to boolean", val, val)

	default:
		// If type is empty or custom (like "map" / "object"), keep original underlying type
		return val, nil
	}
}

func ParseTemplates(yaml *entities.HymlDocument) (map[string]entities.Template, error) {
	templates := make(map[string]entities.Template)

	extractedLocalTemplates := extractLocalTemplates(yaml)
	extractedExternalTemplates := extractExternalTemplates(yaml)

	parsedLocalTemplates, err := parseLocalTemplates(extractedLocalTemplates)
	if err != nil {
		return templates, err
	}

	for _, item := range parsedLocalTemplates {
		templates[item.Name] = item
	}

	parsedExternalTemplates, err := parseExternalTemplates(extractedExternalTemplates)
	if err != nil {
		return templates, err
	}

	for _, item := range parsedExternalTemplates {
		templates[item.Name] = item
	}

	return templates, nil
}

func ParseJsons(yaml *entities.HymlDocument) (map[string]entities.Json, error) {
	jsons := make(map[string]entities.Json)

	extractedJsonReferences := extractJsonReferences(yaml)

	for i, item := range extractedJsonReferences {
		jsonNameVal, exists := item["json"]
		if !exists {
			return nil, fmt.Errorf("item at index %d is missing the required 'json' key", i)
		}
		json, ok := jsonNameVal.(string)
		if !ok {
			return nil, fmt.Errorf("item at index %d has a non-string 'json' key", i)
		}
		fileNameVal, exists := item["file"]
		if !exists {
			return nil, fmt.Errorf("item at index %d is missing the required 'file' key", i)
		}
		file, ok := fileNameVal.(string)
		if !ok {
			return nil, fmt.Errorf("item at index %d has a non-string 'json' key", i)
		}
		jsons[json] = entities.Json{
			Name:    json,
			Path:    file,
			Content: fileReader.ReadJson(file),
		}
	}

	return jsons, nil
}

func parseLocalTemplates(extractedLocalTemplates []map[string]any) ([]entities.Template, error) {
	templates := []entities.Template{}
	for i, item := range extractedLocalTemplates {
		// 1. Extract and validate Name (from "template" key)
		templateNameVal, exists := item["template"]
		if !exists {
			return nil, fmt.Errorf("item at index %d is missing the required 'template' key", i)
		}

		name, ok := templateNameVal.(string)
		if !ok {
			return nil, fmt.Errorf("item at index %d has a non-string 'template' key", i)
		}

		// 2. Extract Content field
		contentVal, hasContent := item["content"]
		if !hasContent {
			return nil, fmt.Errorf("item at index %d ('%s') is missing the required 'content' key", i, name)
		}

		// 3. Determine Locked status (Opposite of enable-overwrite)
		locked := false
		if overwriteVal, hasOverwrite := item["enable-overwrite"]; hasOverwrite {
			// If explicitly a boolean and explicitly false, it is locked
			if overwriteBool, isBool := overwriteVal.(bool); isBool && !overwriteBool {
				locked = true
			}
		}

		// Initialize our entity instance with the baseline data
		tmpl := entities.Template{
			Name:   name,
			Locked: locked,
		}

		tmpl.Content = contentVal.(map[string]any)

		templates = append(templates, tmpl)
	}
	return templates, nil
}

func parseExternalTemplates(extractedExternalTemplates []map[string]any) ([]entities.Template, error) {
	templates := []entities.Template{}

	for i, item := range extractedExternalTemplates {
		// 1. Extract and validate Name (from "template" key)
		templateNameVal, exists := item["template"]
		if !exists {
			return nil, fmt.Errorf("item at index %d is missing the required 'template' key", i)
		}

		name, ok := templateNameVal.(string)
		if !ok {
			return nil, fmt.Errorf("item at index %d has a non-string 'template' key", i)
		}

		// 2. Extract Content field
		contentVal, hasContent := item["content"]
		if !hasContent {
			return nil, fmt.Errorf("item at index %d ('%s') is missing the required path in 'content' key", i, name)
		}

		// 3. Determine Locked status (Opposite of enable-overwrite)
		locked := false
		if overwriteVal, hasOverwrite := item["enable-overwrite"]; hasOverwrite {
			// If explicitly a boolean and explicitly false, it is locked
			if overwriteBool, isBool := overwriteVal.(bool); isBool && !overwriteBool {
				locked = true
			}
		}

		// Initialize our entity instance with the baseline data
		tmpl := entities.Template{
			Name:   name,
			Locked: locked,
		}

		// It's a file reference (e.g., "examples/car.hyml")
		tmpl.ContentPath = contentVal.(string)
		referencedFile := fileReader.ReadPartialYaml(contentVal.(string))

		tmpl.Content = referencedFile.Content
		tmpl.Args, _ = parseArgs(referencedFile.Args)

		templates = append(templates, tmpl)
	}
	return templates, nil
}

func parseArgs(rawArgs entities.RawArgs) (map[string]entities.Arg, error) {
	arguments := make(map[string]entities.Arg)

	for i, item := range rawArgs {
		// 1. Extract Name
		argNameVal, exists := item["name"]
		if !exists {
			return nil, fmt.Errorf("item at index %d is missing the required 'name' key", i)
		}
		name, ok := argNameVal.(string)
		if !ok {
			return nil, fmt.Errorf("item at index %d has a non-string 'name' key", i)
		}

		rawDefault, _ := item["default"]

		rawValue, _ := item["value"]

		// 3. Extract Type string
		argType := ""
		if typeVal, hasType := item["type"]; hasType {
			if tStr, ok := typeVal.(string); ok {
				argType = tStr
			}
		} else {
			fmt.Println("No encontro el tipo")
		}

		parsedDefault, _ := castValueToType(rawDefault, argType)

		parsedValue, _ := castValueToType(rawValue, argType)

		arguments[name] = entities.Arg{
			Name:    name,
			Type:    argType,
			Value:   parsedValue,
			Default: parsedDefault,
		}

	}

	return arguments, nil
}

func extractJsonReferences(yaml *entities.HymlDocument) []map[string]any {
	var jsons []map[string]any

	for _, item := range yaml.Def {
		// 1. Check if the "file" key exists
		fileVal, hasFile := item["file"]

		// 2. Check if the "json" key exists
		_, hasJSONKey := item["json"]

		if hasFile && hasJSONKey {
			// 3. Optional Guardrail: Ensure "file" is a string and actually ends with .json
			if fileStr, isString := fileVal.(string); isString {
				if strings.HasSuffix(fileStr, ".json") {
					jsons = append(jsons, item)
				}
			}
		}
	}
	return jsons
}

func extractVariables(yaml *entities.HymlDocument) []map[string]any {
	var vars []map[string]any

	for _, item := range yaml.Def {
		// 1. Check if the "var" key exists
		_, hasVar := item["var"]

		// 2. Check if the "value" key exists
		_, hasValue := item["value"]

		// If it has both, we treat it as a variable block
		if hasVar && hasValue {
			vars = append(vars, item)
		}
	}
	return vars
}

func extractExternalTemplates(yaml *entities.HymlDocument) []map[string]any {
	var externalTemplates []map[string]any

	for _, item := range yaml.Def {
		// 1. Check if the "template" key exists
		if _, hasTemplate := item["template"]; !hasTemplate {
			continue
		}

		// 2. Check if the "content" key exists
		contentVal, hasContent := item["content"]
		if !hasContent {
			continue
		}

		// 3. Verify that "content" is a string AND ends with ".hyml"
		if contentStr, isString := contentVal.(string); isString {
			if strings.HasSuffix(contentStr, ".hyml") {
				externalTemplates = append(externalTemplates, item)
			}
		}
	}
	return externalTemplates
}

func extractLocalTemplates(yaml *entities.HymlDocument) []map[string]any {
	var localTemplates []map[string]any

	for _, item := range yaml.Def {
		// 1. Check if the "template" key exists
		if _, hasTemplate := item["template"]; !hasTemplate {
			continue
		}

		// 2. Check if the "content" key exists
		contentVal, hasContent := item["content"]
		if !hasContent {
			continue
		}

		// 3. Verify that "content" is actually a map
		_, isMap := contentVal.(map[string]any)
		if isMap {
			localTemplates = append(localTemplates, item)
		}
	}
	return localTemplates
}

func (fileReader FileReader) ReadYaml(filePath string) *entities.HymlDocument {

	data := ReadRawFile(filePath)

	var yamlFile entities.HymlDocument
	err := yaml.Unmarshal(data, &yamlFile)
	if err != nil {
		log.Fatal(err)
	}

	yamlFile.FileName = stringHandler.GetFilenameWithoutExtension(filePath)

	return &yamlFile
}

func (fileReader FileReader) ReadPartialYaml(filePath string) *entities.PartialDocument {

	data := ReadRawFile(filePath)

	var partialFile entities.PartialDocument
	err := yaml.Unmarshal(data, &partialFile)
	if err != nil {
		log.Fatal(err)
	}

	return &partialFile
}

func (fileReader FileReader) ReadJson(filePath string) []byte {
	// 1. Read the raw data (works for JSON or YAML)
	rawData := ReadRawFile(filePath)

	var prettyBuf bytes.Buffer

	// 2. Format the raw JSON bytes directly
	err := json.Indent(&prettyBuf, rawData, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	return prettyBuf.Bytes()
}

// ReadRawFile stays 100% generic. It doesn't care if it's JSON, YAML, or text.
func ReadRawFile(filePath string) []byte {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	// Return the slice directly, no pointer overhead
	return data
}
