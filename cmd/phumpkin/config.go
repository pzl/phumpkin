package main

import (
	"net/http"

	"github.com/pzl/mstk"
	"github.com/pzl/mstk/logger"
	"github.com/pzl/phumpkin/pkg/server"
	"github.com/spf13/pflag"
)

type Cfg struct {
	Listen   string
	PhotoDir string
	ThumbDir string
}

func parseCLI() []server.OptFunc {
	c := mstk.NewConfig("phumpkin", mstk.ConfigFiles("phumpkin"))

	defer func() {
		if r := recover(); r != nil {
			logger.SetFormat(c.Log, false) // flush log if needed
			panic(r)
		}
	}()

	c.SetFlags(func(f *pflag.FlagSet) {
		f.StringP("listen", "l", ":80", "Server listening address")
		f.StringP("PhotoDir", "p", "/photos", "Directory to photo Library")
		f.StringP("ThumbDir", "t", "/thumbs", "Directory to store thumbnails")
		f.AddFlagSet(mstk.CommonFlags())
	})
	err := c.Parse()
	if err != nil {
		panic(err)
	}

	var cfg Cfg
	if err := c.K.Unmarshal("", &cfg); err != nil {
		panic(err)
	}

	opts := []server.OptFunc{
		server.Addr(cfg.Listen),
		server.Log(c.Log),
		server.Photos(cfg.PhotoDir),
		server.Thumbs(cfg.ThumbDir),
		server.Assets(http.FileServer(assets)), // nolint -- assets is generated
	}

	return opts
}
