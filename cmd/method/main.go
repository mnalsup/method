package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"

	"github.com/mnalsup/method/args"
	"github.com/mnalsup/method/cache"
	"github.com/mnalsup/method/request"
)

func main() {
	var definition, cachedDefinition *request.RequestDefinition = &request.RequestDefinition{}, &request.RequestDefinition{}
	fileName := args.ReadRequestFileName()

	err := request.ReadRequestDefinition(fileName, definition)
	if err != nil {
		panic(err.Error())
	}
	err = request.ReadRequestDefinition(cache.GetTempFileName(fileName), cachedDefinition)
	if err == nil {
		cache.MergeTempFileDefinition(definition, cachedDefinition)
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
