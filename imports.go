package xbindata

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

func write_imports(w io.Writer, c *Config, imports ...string) (err error) {
	imports = append(
		imports,
		`bc "github.com/moisespsena-go/xbindata/xbcommon"`,
		`github.com/moisespsena-go/io-common`,
	)

	if c.FileSystem {
		imports = append(
			imports,
			`"github.com/moisespsena-go/xbindata/xbfs"`,
			`fsapi "github.com/moisespsena-go/assetfs/assetfsapi"`,
		)
	}

	var (
		importsMap           = map[string]bool{}
		newImports, excludes []string
	)

	for _, imp := range imports {
		if _, ok := importsMap[imp]; !ok {
			if imp[0] == '-' {
				excludes = append(excludes, imp[0:])
			} else {
				importsMap[imp] = true
			}
		}
	}

	for _, imp := range excludes {
		if _, ok := importsMap[imp]; ok {
			delete(importsMap, imp)
		}
	}

	for imp, _ := range importsMap {
		newImports = append(newImports, imp)
	}

	imports = newImports

	sort.Slice(imports, func(i, j int) bool {
		a, b := imports[i], imports[j]
		ia, ib := strings.IndexByte(a, ' '), strings.IndexByte(b, ' ')
		if ia == ib {
			if len(a) < len(b) {
				return true
			}
			if len(a) == len(b) {
				return a < b
			}
			return false
		} else if ia > 0 && ib < 0 {
			return false
		} else if ib > 0 && ia < 0 {
			return true
		}
		return a[0:ia] < b[0:ib]
	})

	for i, imp := range imports {
		if imp[len(imp)-1] != '"' {
			imp = `"` + imp + `"`
		}
		imports[i] = "\t" + imp
	}
	_, err = fmt.Fprintf(w, "import (\n"+strings.Join(imports, "\n")+"\n)\n")
	return
}
