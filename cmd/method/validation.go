package main

import (
	"fmt"

	"github.com/Jeffail/gabs/v2"
	"github.com/mnalsup/method/authentication"
	"github.com/mnalsup/method/request"
)

func validateAuthenticationHook(result *request.RequestResult, authHook *authentication.AuthenticationHook) (bool, error) {
	if authHook == nil || authHook.RequestPath == "" {
		return false, nil
	}
	for _, v := range authHook.Triggers {
		// Check if configured

		// check if HTTP status mateches the hook
		if v.OnHttpStatus != nil &&
			contains(v.OnHttpStatus, result.Response.StatusCode) {
			return true, nil
		}

		// check if the JSON parsed value matches the hook
		if v.OnJsonValue != nil && result.Body != nil {
			body, err := gabs.ParseJSON(result.Body)
			if err != nil {
				panic(fmt.Errorf("Unable to parse json body for OnJsonValue: %s", err.Error()))
			}
			matchValue, ok := body.Path(v.OnJsonValue.Path).Data().(string)
			if ok {
				match := v.OnJsonValue.Value == matchValue
				if match {
					return true, nil
				}
			}
		}

	}
	return false, nil
}
