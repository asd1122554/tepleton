package cmd

import (
	"fmt"

	"github.com/bgentry/speakeasy"
	"github.com/pkg/errors"
	data "github.com/tepleton/go-data"
	keys "github.com/tepleton/go-keys"
)

const PassLength = 10

func getPassword(prompt string) (string, error) {
	pass, err := speakeasy.Ask(prompt)
	if err != nil {
		return "", err
	}
	if len(pass) < PassLength {
		return "", errors.Errorf("Password must be at least %d characters", PassLength)
	}
	return pass, nil
}

func printInfo(info keys.Info) {
	switch output {
	case "text":
		key, err := data.ToText(info.PubKey)
		if err != nil {
			panic(err) // really shouldn't happen...
		}
		fmt.Printf("%s\t%s\n", info.Name, key)
	case "json":
		json, err := data.ToJSON(info)
		if err != nil {
			panic(err) // really shouldn't happen...
		}
		fmt.Println(string(json))
	}
}

func printInfos(infos keys.Infos) {
	switch output {
	case "text":
		fmt.Println("All keys:")
		for _, i := range infos {
			printInfo(i)
		}
	case "json":
		json, err := data.ToJSON(infos)
		if err != nil {
			panic(err) // really shouldn't happen...
		}
		fmt.Println(string(json))
	}
}