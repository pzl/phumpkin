package main

import (
	"net/http"

	"github.com/pzl/mstk"
	"github.com/pzl/mstk/logger"
	"github.com/pzl/phumpkin/pkg/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type Cfg struct {
	Listen   string
	PhotoDir string
	ThumbDir string
	DataDir  string
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
		f.BoolP("json", "j", false, "output logs in JSON formt")
		f.CountP("verbose", "v", "increased logging. can use multiple times for more")
		f.StringP("listen", "l", ":80", "Server listening address")
		f.StringP("PhotoDir", "p", "/photos", "Directory to photo Library")
		f.StringP("ThumbDir", "t", "/thumbs", "Directory to store thumbnails")
		f.StringP("DataDir", "d", "/data", "Directory to store cache data, and database")
	})

	pflag.Parse()
	if err := setLog(c.Log); err != nil { // sets verbosity and format from pflag
		panic(err)
	}
	if err := c.LoadFlags(); err != nil { // load pflag into koanf
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
		server.DataDir(cfg.DataDir),
		server.Assets(http.FileServer(assets)), // nolint -- assets is generated
	}

	return opts
}

func setLog(log *logrus.Logger) error {
	// set verbosity
	v, err := pflag.CommandLine.GetCount("verbose")
	if err != nil {
		return err
	}
	lvls := []logrus.Level{
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	}
	if v > 3 {
		v = 3
	} else if v < 0 {
		v = 0
	}
	log.SetLevel(lvls[v])

	// set format
	j, err := pflag.CommandLine.GetBool("json")
	if err != nil {
		return err
	}
	logger.SetFormat(log, j)

	return nil
}
