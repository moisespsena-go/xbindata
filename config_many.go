package xbindata

import (
	"fmt"
	"github.com/apex/log"
	"os"
	"path"
	"path/filepath"
	"strings"

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

func (s ManyConfigInputSlice) Items() (r []InputConfig, err error) {
	for j, input := range s {
		if c, err := input.Config(); err != nil {
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
	NamePrefix string `mapstructure:"name_prefix" yaml:"name_prefix"`
	Recursive  bool
	Ignore     IgnoreSlice
	IgnoreGlob IgnoreGlobSlice `mapstructure:"ignore_glob" yaml:"ignore_glob"`
	Pkg        string
}

func (i *ManyConfigInput) GetPkg() string {
	if i.Pkg == "" {
		i.Pkg = path_helpers.StripGoPath(i.Path)
	}
	return i.Pkg
}

func (i *ManyConfigInput) Config() (configs []*InputConfig, err error) {
	if i.Path == "" {
		log.Warnf("input path not set", i.Path)
		return
	}

	if strings.HasPrefix(i.Path, "go:") {
		i.Pkg = i.Path[3:]
		_, i.Path = path_helpers.ResolveGoSrcPath(i.Pkg)
	}

	if _, err = os.Stat(i.Path); err != nil {
		if os.IsNotExist(err) {
			log.Warnf("input %q does not exists", i.Path)
			return nil, nil
		}
		return
	}

	if i.NamePrefix == "_" {
		i.NamePrefix = i.GetPkg()
	} else if strings.Contains(i.NamePrefix, "$PKG") {
		i.NamePrefix = path.Clean(strings.Replace(i.NamePrefix, "$PKG", i.GetPkg(), 1))
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

		func() {
			defer f.Close()
			err = yaml.NewDecoder(f).Decode(&xbinput)
		}()

		if err != nil {
			return
		}

		for _, input := range xbinput.Sources {
			input.NamePrefix = path.Join(i.NamePrefix, input.NamePrefix)

			if input.Prefix != "" {
				if input.Prefix == "_" {
					input.Prefix = input.Path
				}

				input.Prefix = filepath.Join(i.Prefix, input.Prefix)
			}

			input.Path = filepath.Join(i.Path, input.Path)

			if !i.Recursive && input.Recursive {
				input.Recursive = false
			}

			input.IgnoreGlob = append(i.IgnoreGlob, input.IgnoreGlob...)
			input.Ignore = append(i.Ignore, input.Ignore...)

			var cfgs []*InputConfig
			if cfgs, err = input.Config(); err != nil {
				return
			}

			configs = append(configs, cfgs...)
		}
	} else {
		c := &InputConfig{
			Path:       i.Path,
			Recursive:  i.Recursive,
			Prefix:     i.Prefix,
			NamePrefix: i.NamePrefix,
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

		if _, err := os.Stat(filepath.Join(i.Path, ".xbwalk", "main.go")); err == nil {
			c.WalkFunc = i.Walked
		}

		configs = append(configs, c)
	}

	return
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

func (a *ManyConfigCommon) Config() (c *Config, err error) {
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

	if c.Input, err = a.Inputs.Items(); err != nil {
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

func (a *ManyConfigOutlined) Config() (c *Config, err error) {
	if c, err = a.ManyConfigCommon.Config(); err != nil {
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

func (a *ManyConfigOutlined) Translate() (err error) {
	var c *Config
	if c, err = a.Config(); err != nil {
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
