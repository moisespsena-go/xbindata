package xbindata

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/go-errors/errors"
	"github.com/moisespsena-go/xbindata/tempfile"
	"github.com/moisespsena-go/xbindata/walker"
)

func (i ManyConfigInput) Walked(visited *map[string]bool, recursive bool, cb walker.WalkCallback) (err error) {
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

	cmd := exec.Command("go", "build", "-o", exe, filepath.Join(dir, "main.go"))
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

	cmd = exec.Command(exe)
	cmd.Dir = i.Path
	cmd.Env = os.Environ()

	fr, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("pipe stdout failed: %v", err)
	}
	defer fr.Close()
	cmd.Stderr = os.Stderr

	var scanner = bufio.NewScanner(fr)

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("start failed: %v", err)
	}

	var closed bool
	defer func() {
		if !closed {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()

	go func() {
		for scanner.Scan() {
			pth := scanner.Text()
			pth = filepath.Join(i.Path, pth)
			var info os.FileInfo
			if info, err = os.Stat(pth); err != nil {
				return
			}
			if err = cb(walker.FileInfo{FileInfo: info, Path: pth}); err != nil {
				return
			}
		}
	}()

	if err2 := cmd.Wait(); err == nil && err2 != nil {
		err = err2
	}
	closed = true

	return err
}
