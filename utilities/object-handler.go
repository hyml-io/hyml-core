package utilities

import (
	"hyml-core/entities"
	"strings"
)

type ObjectHandler struct{}

var fileReader = FileReader{}

//Object Array Generator - More context logic related

func (objectHandler ObjectHandler) GenerateYamlProperties(yamls []*entities.HymlDocument) []entities.YamlProperty {

	yamlProperties := []entities.YamlProperty{}

	for _, yaml := range yamls {
		yamlProperties = append(yamlProperties, generateProperty("hyml-version", yaml.HymlVersion, yaml.FileName))
		//yamlProperties = append(yamlProperties, generateProperty("Action.Type", yaml.Action.Type, yaml.Header.Name, yaml.Action.CanOverwrite))
		//yamlProperties = append(yamlProperties, generateProperty("Action.ShutdownSignal", yaml.Action.ShutdownSignal, yaml.Header.Name, yaml.Action.CanOverwrite))
		//yamlProperties = append(yamlProperties, generateProperty("Action.Platform.OsFamily", yaml.Action.Platform.OsFamily, yaml.Header.Name, yaml.Action.CanOverwrite))
		//yamlProperties = append(yamlProperties, generateProperty("Action.Platform.PackageInstaller", yaml.Action.Platform.PackageInstaller, yaml.Header.Name, yaml.Action.CanOverwrite))
		//yamlProperties = append(yamlProperties, generateArrayProperty("Action.Platform.InstallationDependencies", yaml.Action.InstallationDependencies, yaml.Header.Name, yaml.Action.CanOverwrite))
		//yamlProperties = append(yamlProperties, generateArrayProperty("Action.InitialInputs", yaml.Action.InitialInputs, yaml.Header.Name, yaml.Action.CanOverwrite))
		yamlProperties = append(yamlProperties, generateDictionaryProperty("def", yaml.Def, yaml.FileName))
		yamlProperties = append(yamlProperties, generateHtmlDictionaryProperty("html", yaml.Html, yaml.FileName))

	}

	return yamlProperties
}

func containsKeyValuePair(values map[string]string, target string) bool {
	for key, value := range values {
		if key == target || value == target {
			return true
		}
	}
	return false
}

//Objectj generator - more context logic related

func generateBoolProperty(name string, value *bool, templateName string, override *bool) entities.YamlProperty {
	yamlProperty := entities.YamlProperty{Name: name, BoolValue: value, TemplateName: templateName}
	if override != nil && (name == "Configuration.BypassSecurity" || name == "Configuration.Containerize") {
		yamlProperty.Sealed = !*override
	}
	return yamlProperty
}

//Objectj generator - more context logic related

func generatePropertyOverwrite(name string, value string, templateName string, override *bool) entities.YamlProperty {
	yamlProperty := generateProperty(name, value, templateName)
	if override != nil {
		yamlProperty.Sealed = !*override
	}
	if strings.Contains(value, "default") {
		yamlProperty.Default = true
	}
	return yamlProperty
}

func generateProperty(name string, value string, templateName string) entities.YamlProperty {
	return entities.YamlProperty{Name: name, Value: value, TemplateName: templateName}
}

//Objectj generator - more context logic related

func generateArrayProperty(name string, values []string, templateName string, override *bool) entities.YamlProperty {
	yamlProperty := entities.YamlProperty{Name: name, Values: values, TemplateName: templateName}

	if override != nil {
		yamlProperty.Sealed = !*override
	}

	if !stringHandler.ContainsString(values, "default") {
		yamlProperty.Default = true
	}
	return yamlProperty
}

//Objectj generator - more context logic related

func generateHtmlDictionaryProperty(name string, values entities.Html, templateName string) entities.YamlProperty {
	yamlProperty := entities.YamlProperty{Name: name, HtmlDictValues: values, TemplateName: templateName}

	return yamlProperty
}

func generateDictionaryProperty(name string, values []map[string]interface{}, templateName string) entities.YamlProperty {
	yamlProperty := entities.YamlProperty{Name: name, DictValues: values, TemplateName: templateName}

	return yamlProperty
}

func generateDictionaryOverwrite(name string, values []map[string]interface{}, templateName string, override *bool) entities.YamlProperty {
	yamlProperty := generateDictionaryProperty(name, values, templateName)
	if override != nil {
		yamlProperty.Sealed = !*override
	}

	return yamlProperty
}
