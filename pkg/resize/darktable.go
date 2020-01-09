package resize

import (
	"bytes"
	"io"
	"os/exec"
	"strconv"
)

func (r *Resizer) runDarktable(j Job) {
	r.log.Trace("starting job")
	defer func() {
		if j.Cancel != nil {
			j.Cancel()
		}
	}()
	maxsize := strconv.Itoa(j.size)

	args := []string{j.source, j.dest, "--width", maxsize, "--height", maxsize, "--hq", strconv.FormatBool(j.hq)}
	if j.xmp != "" { // insert xmp argument if present
		args = append(args, "")
		copy(args[2:], args[1:]) // shift
		args[1] = j.xmp
	}

	r.log.WithField("args", args).Debug("calling darktable-cli")
	cmd := exec.CommandContext(j.ctx, "darktable-cli", args...)

	var buf bytes.Buffer
	sout, o_err := cmd.StdoutPipe()
	if o_err != nil {
		r.log.WithError(o_err).Error("error getting stdout of darktable process")
	}
	serr, e_err := cmd.StderrPipe()
	if e_err != nil {
		r.log.WithError(e_err).Error("error getting stderr of darktable process")
	}

	if err := cmd.Start(); err != nil {
		r.log.WithError(err).Error("error starting darktable-cli process. Unable to perform job")
		return
	}

	if o_err == nil {
		io.Copy(&buf, sout) // nolint
	}
	if e_err == nil {
		io.Copy(&buf, serr) // nolint
	}

	if buf.Len() > 0 {
		r.log.Trace(buf.String())
	}

	err := cmd.Wait()
	if err != nil {
		r.log.WithError(err).Error("darktable-cli exit error")
	} else {
		r.log.Info("darktable-cli exited successfully")
	}
}
