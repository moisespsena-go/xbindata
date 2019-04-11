package xbindata

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/mitchellh/mapstructure"

	"github.com/gobwas/glob"
)

const (
	OutputToProgram = "_"
	OutputToStdout  = "-"
	DefaultDataDir  = "assets"
)

type IgnoreSlice []string

func (s IgnoreSlice) Items() (r []*regexp.Regexp, err error) {
	for j, pattern := range s {
		if regexPattern, err := regexp.Compile(pattern); err != nil {
			return nil, fmt.Errorf("invalid regex pattern #%d (%q): %v", j, pattern, err)
		} else {
			r = append(r, regexPattern)
		}
	}
	return
}

type IgnoreGlobSlice []string

func (s IgnoreGlobSlice) Items() (r []glob.Glob, err error) {
	for j, pattern := range s {
		if globPattern, err := glob.Compile(pattern); err != nil {
			return nil, fmt.Errorf("invalid glob pattern #%d (%q): %v", j, pattern, err)
		} else {
			r = append(r, globPattern)
		}
	}
	return
}

type ManyConfigInputSlice []ManyConfigInput

func (s ManyConfigInputSlice) Items() (r []InputConfig, err error) {
	for j, input := range s {
		if c, err := input.Config(); err != nil {
			return nil, fmt.Errorf("get config from input #%d (%q) failed: %v", j, input, err)
		} else {
			r = append(r, *c)
		}
	}
	return
}

type ManyConfigInput struct {
	Path       string
	Prefix     string
	Recursive  bool
	Ignore     IgnoreSlice
	IgnoreGlob IgnoreGlobSlice `mapstructure:"ignore_glob" yaml:"ignore_glob"`
}

func (i ManyConfigInput) Config() (c *InputConfig, err error) {
	c = &InputConfig{
		Path:      i.Path,
		Recursive: i.Recursive,
		Prefix:    i.Prefix,
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
	return
}

type ManyConfigCommon struct {
	NoAutoLoad bool `mapstructure:"no_auto_load" yaml:"no_auto_load"`
	NoCompress bool `mapstructure:"no_compress" yaml:"no_compress"`
	NoMetadata bool `mapstructure:"no_metadata" yaml:"no_metadata"`
	Mode       uint
	ModTime    int64 `mapstructure:"mod_time" yaml:"mod_time"`
	Ignore     IgnoreSlice
	IgnoreGlob IgnoreGlobSlice `mapstructure:"ignore_glob" yaml:"ignore_glob"`
	Inputs     ManyConfigInputSlice
	Pkg        string
	Output     string
	Prefix     string
	Fs         bool
	Hybrid     bool
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
	c.NoAutoLoad = a.NoAutoLoad
	c.NoCompress = a.NoCompress
	c.NoMetadata = a.NoMetadata
	c.Mode = a.Mode
	c.ModTime = a.ModTime
	c.Prefix = a.Prefix
	c.Hybrid = a.Hybrid

	if a.Output != "" {
		c.Output = a.Output
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
	var hasOutlinedEmbeded bool
	for i, o := range c.Outlined {
		if o.Program {
			if hasOutlinedEmbeded {
				return fmt.Errorf("multiples outlined with embeded enabled")
			}
			hasOutlinedEmbeded = true
		}
		if err = o.Validate(); err != nil {
			return fmt.Errorf("Outlined #d validate failed: %v", i, err)
		}
		c.Outlined[i] = o
	}
	for i, e := range c.Embedded {
		if err = e.Validate(); err != nil {
			return fmt.Errorf("Embeded #d validate failed: %v", i, err)
		}
		c.Embedded[i] = e
	}
	return nil
}

type MapUnmarshaler interface {
	UnmarshalMap(value interface{}) (err error)
}
