package utilities

import (
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

	//fileReader.readDef(yaml.Def)

	if len(yaml.Def) > 0 {

		// Filter the slice
		localTemplates := extractLocalTemplates(yaml)

		// Output the result
		fmt.Printf("Local templates:\n%#v\n", localTemplates)

		localParsedTemplates, err := ParseTemplates(localTemplates)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Local parsed templates:\n%#v\n", localParsedTemplates)

		externalTemplates := extractExternalTemplates(yaml)
		// Output the result
		fmt.Printf("External templates:\n%#v\n", externalTemplates)

		vars := extractVariables(yaml)

		fmt.Printf("Variables:\n%#v\n", vars)

		localParsedVars, err := ParseVars(vars)

		fmt.Printf("Parsed Variables:\n%#v\n", localParsedVars)

		jsons := extractJsonReferences(yaml)

		fmt.Printf("JSON files:\n%#v\n", jsons)

		// 	importInherit := entities.ImportInherit{ParentPath: parentPath, ParentName: parentName}
		// 	yaml.Parent = fileReader.ReadYaml(importInherit.ParentPath)
		// 	newYamlArray := fileReader.ReadAllYamls(importInherit.ParentPath)
		// 	yamlsArray = append(yamlsArray, newYamlArray...)
	}

	yamlsArray = append(yamlsArray, yaml)

	return yamlsArray
}

func ParseVars(extractedMaps []map[string]any) ([]entities.Var, error) {
	var variables []entities.Var

	for i, item := range extractedMaps {
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

		variables = append(variables, entities.Var{
			Name:   name,
			Type:   varType,
			Value:  parsedValue,
			Locked: locked,
		})
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

func ParseTemplates(extractedMaps []map[string]any) ([]entities.Template, error) {
	var templates []entities.Template

	for i, item := range extractedMaps {
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

		// 4. Polymorphically parse Content based on its underlying type
		switch v := contentVal.(type) {
		case string:
			// It's a file reference (e.g., "examples/car.hyml")
			tmpl.ContentPath = v
		case map[string]any:
			// It's an inline nested map structural tree
			tmpl.Content = v
		default:
			return nil, fmt.Errorf("item '%s' has an invalid 'content' type: %T", name, v)
		}

		templates = append(templates, tmpl)
	}

	return templates, nil
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

func (fileReader FileReader) readDef(def entities.Def) {
	fmt.Printf("%+v\n", def)
}

func (fileReader FileReader) ReadYaml(filePath string) *entities.HymlDocument {

	data := ReadRawYaml(filePath)

	var yamlFile entities.HymlDocument
	err := yaml.Unmarshal(*data, &yamlFile)
	if err != nil {
		log.Fatal(err)
	}

	yamlFile.FileName = stringHandler.GetFilenameWithoutExtension(filePath)

	return &yamlFile
}

func ReadRawYaml(filePath string) *[]byte {

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	return &data
}

// func findYamlByName(yamls []*entities.HymlDocument, templateName string) *entities.HymlDocument {
// 	for _, yaml := range yamls {
// 		if yaml.Header.Name == templateName {
// 			return yaml
// 		}
// 	}
// 	return nil
// }
