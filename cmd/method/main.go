package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"

	"github.com/mnalsup/method/cache"
	"github.com/mnalsup/method/request"
)

func main() {
	args := os.Args[1:]

	fileName := args[0]
	fmt.Println(fileName)
	fmt.Println(cache.GetTempFileName(fileName))
	definition, err := request.ReadRequestDefinition(fileName)
	if err != nil {
		panic(err.Error())
	}

	result, err := request.DoMethod(definition)
	if err != nil {
		panic(err.Error())
	}

	out, err := yaml.Marshal(&definition)
	if err != nil {
		panic(err.Error())
	}

	err = ioutil.WriteFile(cache.GetTempFileName(fileName), out, 0666)
	if err != nil {
		panic(err.Error())
	}

	request.PrintRequestResult(result)
}
