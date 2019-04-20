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
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/moisespsena-go/assetfs"
	fsapi "github.com/moisespsena-go/assetfs/assetfsapi"
	"github.com/moisespsena-go/path-helpers"
	"gopkg.in/yaml.v2"
)

type inputStruct struct {
	Path, Prefix, Ns string
}

var (
    loaded          bool
    mu              sync.Mutex
    fs              *assetfs.AssetFileSystem
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
    if loaded {	return }
    mu.Lock()
    defer mu.Unlock()
    if loaded {	return }
	defer func() { loaded = true }()
    fs = assetfs.NewAssetFileSystem()
    
    inputsPath := strings.TrimSuffix(__file__, "_dev.go") + "_inputs.yml"
    f, err := os.Open(inputsPath)
    if err != nil {
    	panic(err)
    }
    defer f.Close()
    
    var inputs []inputStruct
	if err = yaml.NewDecoder(f).Decode(&inputs); err != nil {
		panic(err)
	}

    dir := filepath.Dir(__file__)

	for _, input := range inputs {
		input.Path = filepath.Join(dir, input.Path)
		input.Prefix = filepath.Join(dir, input.Prefix)
		if input.Ns == "" {
			if err = fs.PrependPath(input.Path); err != nil {
				panic(err)
			}
		} else {
			ns := fs.NameSpaceFS(input.Ns)
			if err = ns.PrependPath(input.Path); err != nil {
				panic(err)
			}
		}
	}
}

func FS() *assetfs.AssetFileSystem {
    Load()
    return fs
}
`

	_, err = f.WriteString(data)

	return
}
