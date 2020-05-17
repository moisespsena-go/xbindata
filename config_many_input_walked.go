package xbindata

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-errors/errors"

	"github.com/moisespsena-go/xbindata/tempfile"
	"github.com/moisespsena-go/xbindata/walker"
)

func (i ManyConfigInput) Walked(_ *map[string]bool, prod, _ bool, cb walker.WalkCallback) (err error) {
	dir := filepath.Join(i.Path, ".xbwalk")
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s: %v", dir, err)
		}
	}()

	var suffix, exe string
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}

	if exe, err = tempfile.TempFile("", "xbindata-walk", suffix); err != nil {
		return
	}

	cmd := exec.Command("go", "build", "-tags", "dev", "-o", exe, filepath.Join(dir, "main.go"))
	cmd.Dir = i.Path
	cmd.Env = EnvS(map[string]string{
		"GOOS":   runtime.GOOS,
		"GOARCH": runtime.GOARCH,
	})
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err = cmd.Run(); err != nil {
		err = errors.New("build failed: " + err.Error())
		return
	}

	defer os.Remove(exe)

	var args []string
	if prod {
		args = append(args, "prod")
	}
	cmd = exec.Command(exe, args...)
	cmd.Dir = i.Path
	cmd.Env = os.Environ()

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err = cmd.Run(); err != nil {
		return
	}

	for _, pth := range strings.Split(out.String(), "\n") {
		pth = strings.TrimSpace(pth)
		pth = filepath.Join(i.Path, pth)
		var info os.FileInfo
		if info, err = os.Stat(pth); err != nil {
			return
		}
		if err = cb(walker.FileInfo{FileInfo: info, Path: pth}); err != nil {
			return
		}
	}

	return err
}
