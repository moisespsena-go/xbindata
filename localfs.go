package xbindata

import (
	"os"
	"strings"
)

const tagDev = "dev"

func localFs(c *Config) (err error) {
	var (
		f   *os.File
		pth string
	)
	if c.Outlined {
		pth = strings.TrimSuffix(c.OutlinedApi, ".go") + "_dev.go"
	}

	if f, err = os.Create(pth); err != nil {
		return
	}

	defer f.Close()

	if _, err = f.WriteString("// +build " + tagDev + "\n\npackage " + c.Package + "\n"); err != nil {
		return
	}
	var data string

	data += `
import (
	"bufio"
	"os"
	"path"
	"strings"
	"sync"

	path_helpers "github.com/moisespsena-go/path-helpers"
	"gopkg.in/yaml.v2"

	"github.com/moisespsena-go/assetfs"
	fsapi "github.com/moisespsena-go/assetfs/assetfsapi"

	bc "github.com/moisespsena-go/xbindata/xbcommon"
	"github.com/moisespsena-go/xbindata/xbfs"
)

var (
	loaded          bool
	mu              sync.Mutex
	fs              fsapi.Interface
	fsLoadCallbacks []func(fs fsapi.Interface)

	__file__ = path_helpers.GetCalledFile(true)
)

func OnFsLoad(cb ...func(fs fsapi.Interface)) {
	fsLoadCallbacks = append(fsLoadCallbacks, cb...)
}

func callFsLoadCallbacks() {
	for _, f := range fsLoadCallbacks {
		f(fs)
	}
}

func Load() {
	if loaded {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	if loaded {
		return
	}
	defer func() { loaded = true }()

	pathsF, err := os.Open(strings.TrimSuffix(__file__, "_dev.go") + "_toc_paths.yml")
	if err != nil {
		panic(err)
	}
	defer pathsF.Close()

	namesF, err := os.Open(strings.TrimSuffix(__file__, "_dev.go") + "_toc_names.yml")
	if err != nil {
		panic(err)
	}
	defer namesF.Close()

	pathScanner := bufio.NewScanner(pathsF)
	nameScanner := bufio.NewScanner(namesF)

	// skip first line
	pathScanner.Scan()
	nameScanner.Scan()

	var assets []bc.Asset

	for pathScanner.Scan() && nameScanner.Scan() {
		var (
			nameS []struct{ Name string }
			pathS []string
			pth   = pathScanner.Text()
			name  = nameScanner.Text()
		)
		yaml.Unmarshal([]byte(pth), &pathS)
		yaml.Unmarshal([]byte(name), &nameS)
		pth, name = path.Join(path.Dir(__file__), pathS[0]), nameS[0].Name

		if info, err := os.Stat(pth); err != nil {
			panic(err)
		} else {
			localAsset := assetfs.NewRealFileInfo(fsapi.OsFileInfoToBasic(name, info), pth)
			assets = append(assets, &bc.LocalFile{RealFileInfo: localAsset})
		}
	}

	fs = xbfs.NewFileSystem(bc.NewTree(assets...).Root())
}

func FS() fsapi.Interface {
	Load()
	return fs
}
`

	_, err = f.WriteString(data)

	return
}
