package xbindata

import (
	"bufio"
	"fmt"
	"github.com/moisespsena-go/xbindata/walker"
	"os"
	"os/exec"
	"path/filepath"
)

func (i ManyConfigInput) Walked(visited *map[string]bool, recursive bool, cb walker.WalkCallback) (err error) {
	cmd := exec.Command("go", "run", filepath.Join(i.Path, ".xbwalk", "main.go"))
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

	cmd.Wait()
	closed = true

	return err
}
