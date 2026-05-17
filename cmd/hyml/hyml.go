package main

import (
	"fmt"
	"hyml-core/utilities"
	"os"
)

var fileReader = utilities.FileReader{}
var objectHandler = utilities.ObjectHandler{}

func main() {

	filePath := os.Args[1]

	yamls := fileReader.ReadAllYamls(filePath)

	generalProperties := objectHandler.GenerateYamlProperties(yamls)

	fmt.Printf("---\n---\n---Vienen las propiedades\n")

	fmt.Printf("%+v\n", generalProperties)

}
