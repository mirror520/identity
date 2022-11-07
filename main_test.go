package main

import (
	"fmt"
	"testing"

	"github.com/jinzhu/configor"

	"github.com/mirror520/jinte/model"
)

func TestLoadConfig(t *testing.T) {
	configor.Load(&model.Config, "config.example.yaml")
	config := model.Config
	fmt.Println(config)
}
