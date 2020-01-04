package photos

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/DAddYE/vips"
)

func Resize(src string, dest string, px int) error {
	var in []byte
	var err error

	ext := strings.ToLower(path.Ext(src))
	switch ext {
	case ".jpg", ".jpeg":
		in, err = fromjpg(src)
	default:
		in, err = fromexif(src)
	}
	if err != nil {
		return err
	}

	buf, err := vips.Resize(in, vips.Options{
		Width:   px,
		Height:  px,
		Quality: 60,
		Crop:    false,
		Gravity: vips.CENTRE,
	})
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(dest, buf, 0644)
}

func fromjpg(src string) ([]byte, error) { return ioutil.ReadFile(src) }

func fromexif(src string) ([]byte, error) {
	c := exec.Command("exiftool", "-b", "-ThumbnailImage", src)
	sout, err := c.StdoutPipe()
	if err != nil {
		return nil, err
	}
	serr, err := c.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := c.Start(); err != nil {
		return nil, err
	}
	output, err := ioutil.ReadAll(sout)
	if err != nil {
		return nil, err
	}
	errput, err := ioutil.ReadAll(serr)
	if err != nil {
		return nil, err
	}

	if err := c.Wait(); err != nil {
		return nil, err
	}

	if len(errput) > 0 {
		return nil, errors.New(string(errput))
	}

	return output, nil
}
