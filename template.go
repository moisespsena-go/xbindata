package xbindata

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

func (i *ManyConfigInput) format(ctx context.Context, key, value string, envs ...map[string]string) (v string, err error) {
	defer func() {
		if err != nil {
			err = errors.New(fmt.Sprint("format{"+key+"}:", err.Error()))
		}
	}()

	env := Env(append([]map[string]string{ContextEnv(ctx)}, envs...)...)
	if strings.ContainsRune(value, '{') {
		var t *template.Template
		if t, err = template.New("").Parse(key); err != nil {
			return
		}
		var buf bytes.Buffer
		if err = t.Execute(&buf, map[string]interface{}{
			"pclean":  path.Clean,
			"pjoin":   path.Join,
			"fpclean": filepath.Clean,
			"fpjoin":  filepath.Join,
			"Env":     env,
			"PKG":     i.GetPkg(),
		}); err != nil {
			return
		}
		v = buf.String()
	} else {
		v = value
	}
	return
}
