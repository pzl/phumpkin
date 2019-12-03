// +build ignore

package main

import (
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() { // nolint
	fs := http.Dir("../../frontend/dist")

	err := vfsgen.Generate(fs, vfsgen.Options{
		Filename:     "assets.go",
		VariableName: "assets",
	})
	if err != nil {
		panic(err)
	}
}
