package test

import (
	"fmt"
	"testing"
	"yfig"
)

func TestYml(t *testing.T) {
	file, err := yfig.LoadYamlFile("E:\\git_code\\yfig\\test\\test.yaml")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(file)
}
