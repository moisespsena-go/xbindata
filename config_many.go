package xbindata

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/gobwas/glob"
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
	Recursive  bool
	Ignore     IgnoreSlice     `mapstructure:""`
	IgnoreGlob IgnoreGlobSlice `mapstructure:"ignore_glob" yaml:"ignore_glob"`
}

func (i ManyConfigInput) Config() (c *InputConfig, err error) {
	c = &InputConfig{
		Path:      i.Path,
		Recursive: i.Recursive,
	}

	if c.IgnoreGlob, err = i.IgnoreGlob.Items(); err != nil {
		return nil, err
	}
	if c.Ignore, err = i.Ignore.Items(); err != nil {
		return nil, err
	}
	return
}

type ManyConfigEmbedded struct {
}

type ManyConfigOutlined struct {
	Pkg        string
	Api        string
	Output     string
	Prefix     string
	Fs         bool
	Hybrid     bool
	Embeded    bool
	NoAutoLoad bool `mapstructure:"no_auto_load" yaml:"no_auto_load"`
	NoCompress bool `mapstructure:"no_compress" yaml:"no_compress"`
	NoMetadata bool `mapstructure:"no_metadata" yaml:"no_metadata"`
	Mode       uint
	ModTime    int64 `mapstructure:"mod_time" yaml:"mod_time"`
	Ignore     IgnoreSlice
	IgnoreGlob IgnoreGlobSlice `mapstructure:"ignore_glob" yaml:"ignore_glob"`
	Inputs     ManyConfigInputSlice
}

func (a *ManyConfigOutlined) Validate() (err error) {
	if a.Pkg == "" {
		a.Pkg = "main"
	}
	if a.Pkg == "main" {
		if a.Api == "" {
			a.Api = "assets.go"
		}

		if len(a.Inputs) == 0 {
			a.Inputs = append(a.Inputs, ManyConfigInput{
				Path:       "assets",
				Recursive:  true,
				IgnoreGlob: IgnoreGlobSlice{".*", "*.swp"},
			})
		}
	} else {
		dir := filepath.FromSlash(a.Pkg)

		if a.Api == "" {
			a.Api = filepath.Join(dir, "assets.go")
		}

		if len(a.Inputs) == 0 {
			a.Inputs = append(a.Inputs, ManyConfigInput{
				Path:       filepath.Join(dir, "assets"),
				Recursive:  true,
				IgnoreGlob: IgnoreGlobSlice{".*", "*.swp"},
			})
		}
	}

	if a.Prefix == "." {
		a.Prefix = a.Inputs[0].Path
	}

	return nil
}

func (a *ManyConfigOutlined) Config() (c *Config, err error) {
	c = NewConfig()
	c.Outlined = true
	c.Package = a.Pkg
	c.OutlinedApi = a.Api
	c.FileSystem = a.Fs
	c.Hybrid = a.Hybrid
	c.NoAutoLoad = a.NoAutoLoad
	c.NoCompress = a.NoCompress
	c.NoMetadata = a.NoMetadata
	c.Mode = a.Mode
	c.ModTime = a.ModTime
	c.Prefix = a.Prefix
	c.Output = ""

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

func (c *ManyConfig) Validate() (err error) {
	var hasOutlinedEmbeded bool
	for i, o := range c.Outlined {
		if o.Embeded {
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
	return nil
}
