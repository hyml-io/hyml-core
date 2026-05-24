package main

import (
	"fmt"
	"hyml-core/utilities"
	"os"
)

var fileReader = utilities.FileReader{}

func main() {

	filePath := os.Args[1]

	yamls := fileReader.ReadAllYamls(filePath)

	fmt.Printf("\nLos Yamls: %+v\n", yamls)
}
