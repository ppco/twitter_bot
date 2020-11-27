package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

const (
	credsFile = "creds.yml"
)

type creds struct {
	ConsumerKey    string `yaml:"ConsumerKey"`
	ConsumerSecret string `yaml:"ConsumerSecret"`
	AccessToken    string `yaml:"AccessToken"`
	AccessSecret   string `yaml:"AccessSecret"`
}

func getCreds() *creds {
	c := &creds{}

	yamlFile, err := ioutil.ReadFile(credsFile)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	return c
}
