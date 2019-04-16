package xbindata

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func gitIgnore(dir string, lines ...string) (err error) {
	pth := filepath.Join(dir, ".gitignore")
	if f, err := os.Open(pth); err == nil {
		s, err := f.Stat()
		if err != nil {
			f.Close()
			return err
		}
		var has = map[string]bool{}

		err = func() error {
			defer f.Close()
			var r = bufio.NewReader(f)
			var line []byte
			for {
				if line, _, err = r.ReadLine(); err == nil {
					for _, l := range lines {
						if string(line) == l {
							has[l] = true
							break
						}
					}
				} else if err == io.EOF {
					break
				} else {
					return fmt.Errorf("read %q gitignore failed: `%v`\n", pth, err)
				}
			}
			return nil
		}()
		if err != nil {
			return err
		}

		if len(has) < len(lines) {
			if f, err = os.OpenFile(pth, os.O_APPEND|os.O_CREATE|os.O_WRONLY, s.Mode()); err != nil {
				return err
			}
			func() {
				defer f.Close()
				for _, line := range lines {
					if _, ok := has[line]; !ok {
						if _, err = f.WriteString(line + "\n"); err != nil {
							err = fmt.Errorf("add `%s` to gitignore failed: %v", line, err)
							return
						}
					}
				}
				return
			}()
			if err != nil {
				return err
			}
		}
	} else if os.IsNotExist(err) {
		if err = safefileWriteFile(pth, []byte(strings.Join(lines, "\n")+"\n"), 0); err != nil {
			return err
		}
	}
	return
}
