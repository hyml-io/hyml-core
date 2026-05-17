package utilities

import (
	"fmt"
	"hyml-core/entities"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type FileReader struct {
}

var stringHandler = StringHandler{}

func (fileReader FileReader) ReadAllYamls(path string) []*entities.HymlDocument {

	yamlsArray := make([]*entities.HymlDocument, 0)

	yaml := fileReader.ReadYaml(path)

	fileReader.readDef(yaml.Def)

	// if len(yaml.Def) > 0 {
	// 	importInherit := entities.ImportInherit{ParentPath: parentPath, ParentName: parentName}
	// 	yaml.Parent = fileReader.ReadYaml(importInherit.ParentPath)
	// 	newYamlArray := fileReader.ReadAllYamls(importInherit.ParentPath)
	// 	yamlsArray = append(yamlsArray, newYamlArray...)
	// }

	yamlsArray = append(yamlsArray, yaml)

	return yamlsArray
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
