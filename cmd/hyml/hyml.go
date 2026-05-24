package main

import (
	"hyml-core/utilities"
	"os"
)

var fileReader = utilities.FileReader{}

func main() {

	filePath := os.Args[1]

	fileReader.ReadAllYamls(filePath)
}
