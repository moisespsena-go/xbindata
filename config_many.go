package xbindata

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-errors/errors"

	"github.com/apex/log"

	"gopkg.in/yaml.v2"

	"github.com/moisespsena-go/path-helpers"
	"github.com/moisespsena-go/xbindata/ignore"

	"github.com/mitchellh/mapstructure"
)

const (
	OutputToProgram = "_"
	OutputToStdout  = "-"
	DefaultDataDir  = "assets"
)

type (
	IgnoreSlice     = ignore.IgnoreSlice
	IgnoreGlobSlice = ignore.IgnoreGlobSlice

	ManyConfigInputSlice []ManyConfigInput
)

func (s ManyConfigInputSlice) Items(ctx context.Context) (r []InputConfig, err error) {
	for j, input := range s {
		if c, err := input.Config(ContextWithInputKey(ctx, "#"+strconv.Itoa(j))); err != nil {
			return nil, fmt.Errorf("get config from input #%d (%q) failed: %v", j, input, err)
		} else {
			for _, c := range c {
				r = append(r, *c)
			}
		}
	}
	return
}

type ManyConfigInput struct {
	Path       string
	Prefix     string
	NameSpace  string `mapstructure:"ns" yaml:"ns"`
	Recursive  bool
	Ignore     IgnoreSlice
	IgnoreGlob IgnoreGlobSlice `mapstructure:"ignore_glob" yaml:"ignore_glob"`
	Pkg        string
}

func (i *ManyConfigInput) UnmarshalMap(value interface{}) (err error) {
	var ns struct{ Ns string }
	if err = mapstructure.Decode(value, i); err == nil {
		if i.NameSpace == "" {
			if err = mapstructure.Decode(value, &ns); err == nil {
				i.NameSpace = ns.Ns
			}
		}
	}
	return
}

func (i *ManyConfigInput) pkgSetup() {
	if i.Pkg == "" {
		pth := i.Path
		if !filepath.IsAbs(pth) {
			var err error
			if pth, err = filepath.Abs(pth); err != nil {
				panic(fmt.Errorf("abs path of %q failed: %v", pth, err))
			}
		}
		i.Pkg = path_helpers.StripGoPath(pth)
	}
}

func (i ManyConfigInput) Config(ctx context.Context) (configs []*InputConfig, err error) {
	if i.Path == "" {
		log.Warnf("input path not set", i.Path)
		return
	}

	if key := InputKey(ctx, "[%s]:"); key != "" {
		defer func() {
			if err != nil {
				err = errors.New(fmt.Sprint(key, err.Error()))
			}
		}()
	}

	if strings.HasPrefix(i.Path, "go:") {
		i.Pkg = i.Path[3:]
		_, i.Path = path_helpers.ResolveGoSrcPath(i.Pkg)
	}

	if i.Path, err = i.format(ctx, "path", i.Path); err != nil {
		return
	}

	if _, err = os.Stat(i.Path); err != nil {
		if os.IsNotExist(err) {
			log.Warnf("input %q does not exists", i.Path)
			return nil, nil
		}
		return
	}

	i.pkgSetup()

	if i.NameSpace != "" {
		if i.NameSpace == "_" {
			i.NameSpace = i.Pkg
		} else if strings.Contains(i.NameSpace, "$PKG") {
			i.NameSpace = strings.Replace(i.NameSpace, "$PKG", i.Pkg, 1)
		}

		if i.NameSpace, err = i.format(ctx, "name_space", i.NameSpace); err != nil {
			return
		}

		i.NameSpace = path.Clean(i.NameSpace)
	}

	xbinputFile := filepath.Join(i.Path, ".xbinputs.yml")
	if _, err = os.Stat(xbinputFile); err == nil {
		if i.Prefix == "_" {
			i.Prefix = i.Path
		}

		var (
			xbinput struct{ Sources []ManyConfigInput }
			f       *os.File
		)
		if f, err = os.Open(xbinputFile); err != nil {
			return
		}

		ctx := ContextWithInputKey(ctx, xbinputFile)

		func() {
			defer f.Close()
			err = yaml.NewDecoder(f).Decode(&xbinput)
		}()

		if err != nil {
			return
		}

		for _, input := range xbinput.Sources {
			if input.NameSpace != "" && i.NameSpace != "" {
				input.NameSpace = i.NameSpace + "/" + input.NameSpace
			}

			if input.Path[0] == '/' {
				// example: /{{.Env.GOROOT}}/a/b
				if input.Path[1] == '{' {
					input.Path = input.Path[1:]
				}
				input.Path, _ = input.format(ctx, "Path", input.Path)

				if input.Prefix == "_" {
					input.Prefix = input.Path
				}
			} else {
				if input.Prefix != "" {
					if input.Prefix == "_" {
						input.Prefix = input.Path
					}

					input.Prefix = filepath.Join(i.Prefix, input.Prefix)
				}

				input.Path = filepath.Join(i.Path, input.Path)
			}

			if !i.Recursive && input.Recursive {
				input.Recursive = false
			}

			input.IgnoreGlob = append(i.IgnoreGlob, input.IgnoreGlob...)
			input.Ignore = append(i.Ignore, input.Ignore...)

			var cfgs []*InputConfig
			if cfgs, err = input.Config(ctx); err != nil {
				return
			}

			configs = append(configs, cfgs...)
		}
		return
	}

	c := &InputConfig{
		Path:      i.Path,
		Recursive: i.Recursive,
		Prefix:    i.Prefix,
		NameSpace: i.NameSpace,
	}

	if i.Prefix == "_" {
		c.Prefix = i.Path
	}

	if c.IgnoreGlob, err = i.IgnoreGlob.Items(); err != nil {
		return nil, err
	}
	if c.Ignore, err = i.Ignore.Items(); err != nil {
		return nil, err
	}

	walkedPath := filepath.Join(i.Path, ".xbwalk", "main.go")
	if _, err := os.Stat(walkedPath); err == nil {
		c.WalkFunc = i.Walked
	}

	configs = append(configs, c)
	return
}

type ManyConfigCommonDefaultInput struct {
	Prefix    string
	NameSpace string `mapstructure:"name_space" yaml:"name_space"`
	Recursive bool
}

func (i *ManyConfigCommonDefaultInput) UnmarshalMap(value interface{}) (err error) {
	var ns struct{ Ns string }
	if err = mapstructure.Decode(value, i); err == nil {
		if i.NameSpace == "" {
			if err = mapstructure.Decode(value, &ns); err == nil {
				i.NameSpace = ns.Ns
			}
		}
	}
	return
}

type ManyConfigCommonDefault struct {
	Input ManyConfigCommonDefaultInput
}

func (d *ManyConfigCommonDefault) UnmarshalMap(value interface{}) (err error) {
	return mapstructure.Decode(value, d)
}

type ManyConfigCommon struct {
	Disabled        bool
	NoAutoLoad      bool `mapstructure:"no_auto_load" yaml:"no_auto_load"`
	NoCompress      bool `mapstructure:"no_compress" yaml:"no_compress"`
	NoMetadata      bool `mapstructure:"no_metadata" yaml:"no_metadata"`
	NoMemCopy       bool `mapstructure:"no_mem_copy" yaml:"no_mem_copy"`
	Mode            uint
	ModTime         int64 `mapstructure:"mod_time" yaml:"mod_time"`
	Ignore          IgnoreSlice
	IgnoreGlob      IgnoreGlobSlice `mapstructure:"ignore_glob" yaml:"ignore_glob"`
	Inputs          ManyConfigInputSlice
	Pkg             string
	Output          string
	Prefix          string
	Hybrid          bool
	Fs              bool
	FsLoadCallbacks []string `mapstructure:"fs_load_callbacks" yaml:"fs_load_callbacks"`
	Default         ManyConfigCommonDefault
}

func (a *ManyConfigCommon) Validate() (err error) {
	if a.Pkg == "" {
		a.Pkg = "main"
	}
	if a.Pkg == "main" {
		if len(a.Inputs) == 0 {
			a.Inputs = append(a.Inputs, ManyConfigInput{
				Path:       "assets",
				Recursive:  true,
				IgnoreGlob: IgnoreGlobSlice{".*", "*.swp"},
			})
		}
	} else {
		if len(a.Inputs) == 0 {
			a.Inputs = append(a.Inputs, ManyConfigInput{
				Path:       filepath.Join(filepath.FromSlash(a.Pkg), DefaultDataDir),
				Recursive:  true,
				IgnoreGlob: IgnoreGlobSlice{".*", "*.swp"},
			})
		}
	}

	if a.Prefix == "_" {
		a.Prefix = a.Inputs[0].Path
	}

	return nil
}

func (a *ManyConfigCommon) Config(ctx context.Context) (c *Config, err error) {
	c = NewConfig()
	c.Package = a.Pkg
	c.FileSystem = a.Fs
	c.FileSystemLoadCallbacks = a.FsLoadCallbacks
	c.NoAutoLoad = a.NoAutoLoad
	c.NoCompress = a.NoCompress
	c.NoMetadata = a.NoMetadata
	c.NoMemCopy = a.NoMemCopy
	c.Mode = a.Mode
	c.ModTime = a.ModTime
	c.Prefix = a.Prefix
	c.Hybrid = a.Hybrid

	if a.Output != "" {
		c.Output = a.Output
	}

	if len(a.FsLoadCallbacks) > 0 {
		var cwd, _ = os.Getwd()
		gph := path_helpers.PkgFromPath(cwd)

		for i, cb := range a.FsLoadCallbacks {
			if cb[0] == '.' {
				cb = path.Join(gph, cb[1:])
				c.FileSystemLoadCallbacks[i] = cb
			}
		}
	}

	if c.IgnoreGlob, err = a.IgnoreGlob.Items(); err != nil {
		return nil, err
	}
	if c.Ignore, err = a.Ignore.Items(); err != nil {
		return nil, err
	}

	for i, input := range a.Inputs {
		if a.Default.Input.Prefix != "" && input.Prefix == "" {
			input.Prefix = a.Default.Input.Prefix
		}
		if a.Default.Input.NameSpace != "" && input.NameSpace == "" {
			input.NameSpace = a.Default.Input.NameSpace
		}
		if a.Default.Input.Recursive {
			input.Recursive = true
		}
		a.Inputs[i] = input
	}

	if c.Input, err = a.Inputs.Items(ctx); err != nil {
		return nil, err
	}

	return
}

func (a *ManyConfigCommon) UnmarshalMap(value interface{}) (err error) {
	return mapstructure.Decode(value, a)
}

type ManyConfigEmbedded struct {
	ManyConfigCommon
}

func (a *ManyConfigEmbedded) Validate() (err error) {
	if err = a.ManyConfigCommon.Validate(); err != nil {
		return
	}
	if a.Output == "" {
		if a.Pkg == "main" {
			a.Output = "assets.go"
		} else {
			a.Output = filepath.Join(filepath.FromSlash(a.Pkg), "assets.go")
		}
	}

	return nil
}

func (a *ManyConfigEmbedded) UnmarshalMap(value interface{}) (err error) {
	return a.ManyConfigCommon.UnmarshalMap(value)
}

type ManyConfigOutlined struct {
	ManyConfigCommon
	Api     string
	Program bool
}

func (a *ManyConfigOutlined) Validate() (err error) {
	if err = a.ManyConfigCommon.Validate(); err != nil {
		return
	}
	if a.Pkg == "main" {
		if a.Api == "" {
			a.Api = "assets.go"
		}
	} else if a.Api == "" {
		a.Api = filepath.Join(filepath.FromSlash(a.Pkg), "assets.go")
	}

	return nil
}

func (a *ManyConfigOutlined) Config(ctx context.Context) (c *Config, err error) {
	if c, err = a.ManyConfigCommon.Config(ctx); err != nil {
		return
	}
	c.Outlined = true
	c.OutlinedApi = a.Api
	c.OutlinedProgram = a.Program
	c.Output = ""

	if a.Output == "" && !a.Program {
		c.Output = filepath.Join(c.OutlinedOutputDir, filepath.FromSlash(a.Pkg)+".xb")
	} else {
		c.Output = a.Output
	}
	return
}

func (a *ManyConfigOutlined) UnmarshalMap(value interface{}) (err error) {
	if err = a.ManyConfigCommon.UnmarshalMap(value); err != nil {
		return
	}
	err = mapstructure.Decode(value, a)
	return
}

func (a *ManyConfigOutlined) Translate(ctx context.Context) (count int, err error) {
	var c *Config
	if c, err = a.Config(ctx); err != nil {
		return
	}
	return Translate(c)
}

type ManyConfig struct {
	Embedded []ManyConfigEmbedded
	Outlined []ManyConfigOutlined
}

func (c *ManyConfig) InputsRelTo(baseDir string) {
	for _, cfg := range c.Embedded {
		for _, input := range cfg.Inputs {
			if newPth, err := filepath.Rel(baseDir, input.Path); err == nil {
				input.Path = newPth
			}
		}
	}
	for _, cfg := range c.Outlined {
		for _, input := range cfg.Inputs {
			if newPth, err := filepath.Rel(baseDir, input.Path); err == nil {
				input.Path = newPth
			}
		}
	}
}

func (c *ManyConfig) Validate() (err error) {
	var (
		hasOutlinedEmbeded bool
		embedded           []ManyConfigEmbedded
		outlined           []ManyConfigOutlined
	)
	for i, cfg := range c.Outlined {
		if cfg.Disabled {
			continue
		}
		if cfg.Program {
			if hasOutlinedEmbeded {
				return fmt.Errorf("multiples outlined with embeded enabled")
			}
			hasOutlinedEmbeded = true
		}
		if err = cfg.Validate(); err != nil {
			return fmt.Errorf("Outlined #d validate failed: %v", i, err)
		}
		outlined = append(outlined, cfg)
	}
	for i, cfg := range c.Embedded {
		if cfg.Disabled {
			continue
		}
		if err = cfg.Validate(); err != nil {
			return fmt.Errorf("Embeded #d validate failed: %v", i, err)
		}
		embedded = append(embedded, cfg)
	}
	c.Outlined, c.Embedded = outlined, embedded
	return nil
}

type MapUnmarshaler interface {
	UnmarshalMap(value interface{}) (err error)
}
