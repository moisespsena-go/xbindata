// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package xbindata

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/moisespsena-go/xbindata/walker"

	"github.com/gobwas/glob"
)

const DefaultOutput = "assets.go"

// InputConfig defines options on a asset directory to be convert.
type InputConfig struct {
	// Path defines a directory containing asset files to be included
	// in the generated output.
	Path string

	// Recusive defines whether subdirectories of Path
	// should be recursively included in the conversion.
	Recursive bool

	// Ignores any filenames matching the regex pattern specified, e.g.
	// path/to/file.ext will ignore only that file, or \\.gitignore
	// will match any .gitignore file.
	//
	// This parameter can be provided multiple times.
	Ignore []*regexp.Regexp

	// Ignores any filenames matching the glob pattern specified, e.g.
	// path/to/*.ext will ignore only that file, or \\.gitignore
	// will match any .gitignore file.
	// See <https://github.com/gobwas/glob> for more glob info.
	//
	// This parameter can be provided multiple times.
	IgnoreGlob []glob.Glob

	Prefix string

	NameSpace string

	DirReplacesCount int

	WalkFunc func(visited *map[string]bool, prod, recursive bool, cb func(info walker.FileInfo) error) error
}

func (i InputConfig) Walk(visited *map[string]bool, prod bool, cb walker.WalkCallback) (err error) {
	if i.WalkFunc != nil {
		return i.WalkFunc(visited, prod, i.Recursive, i.prepareCb(cb))
	}

	return i.DefaultWalk(visited, i.Recursive, cb)
}

func (i InputConfig) DefaultWalk(visited *map[string]bool, recursive bool, cb walker.WalkCallback) (err error) {
	var pth = i.Path
	w := walker.Walker{Recursive: recursive, VisitedPaths: visited}
	return w.Walk(pth, i.prepareCb(cb))
}

func (i InputConfig) prepareCb(cb walker.WalkCallback) walker.WalkCallback {
	if i.NameSpace != "" {
		old := cb
		cb = func(info walker.FileInfo) error {
			return old(info.SetNamePrefix(i.NameSpace))
		}
	}
	return cb
}

// Config defines a set of options for the asset conversion.
type Config struct {
	// Name of the package to use. Defaults to 'main'.
	Package string

	// Tags specify a set of optional build tags, which should be
	// included in the generated output. The tags are appended to a
	// `// +build` line in the beginning of the output file
	// and must follow the build tags syntax specified by the go tool.
	Tags []string

	// Input defines the directory path, containing all asset files as
	// well as whether to recursively process assets in any sub directories.
	Input []InputConfig

	// Output defines the output file for the generated code.
	// If left empty, this defaults to 'xbindata.go' in the current
	// working directory.
	Output string

	// OutputWriter defines the output writer for the generated code.
	OutputWriter io.Writer

	// Prefix defines a path prefix which should be stripped from all
	// file names when generating the keys in the table of contents.
	// For example, running without the `-prefix` flag, we get:
	//
	// 	$ xbindata /path/to/templates
	// 	go_bindata["/path/to/templates/foo.html"] = _path_to_templates_foo_html
	//
	// Running with the `-prefix` flag, we get:
	//
	// 	$ xbindata -prefix "/path/to/" /path/to/templates/foo.html
	// 	go_bindata["templates/foo.html"] = templates_foo_html
	Prefix string

	// NoMemCopy will alter the way the output file is generated.
	//
	// It will employ a hack that allows us to read the file data directly from
	// the compiled program's `.rodata` section. This ensures that when we call
	// call our generated function, we omit unnecessary mem copies.
	//
	// The downside of this, is that it requires dependencies on the `reflect` and
	// `unsafe` packages. These may be restricted on platforms like AppEngine and
	// thus prevent you from using this mode.
	//
	// Another disadvantage is that the byte slice we create, is strictly read-only.
	// For most use-cases this is not a problem, but if you ever try to alter the
	// returned byte slice, a runtime panic is thrown. Use this mode only on target
	// platforms where memory constraints are an issue.
	//
	// The default behaviour is to use the old code generation method. This
	// prevents the two previously mentioned issues, but will employ at least one
	// extra memcopy and thus increase memory requirements.
	//
	// For instance, consider the following two examples:
	//
	// This would be the default mode, using an extra memcopy but gives a safe
	// implementation without dependencies on `reflect` and `unsafe`:
	//
	// 	func myfile() []byte {
	// 		return []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a}
	// 	}
	//
	// Here is the same functionality, but uses the `.rodata` hack.
	// The byte slice returned from this example can not be written to without
	// generating a runtime error.
	//
	// 	var _myfile = "\x89\x50\x4e\x47\x0d\x0a\x1a"
	//
	// 	func myfile() []byte {
	// 		var empty [0]byte
	// 		sx := (*reflect.StringHeader)(unsafe.Pointer(&_myfile))
	// 		b := empty[:]
	// 		bx := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	// 		bx.Data = sx.Data
	// 		bx.Len = len(_myfile)
	// 		bx.Cap = bx.Len
	// 		return b
	// 	}
	NoMemCopy bool

	// NoCompress means the assets are /not/ GZIP compressed before being turned
	// into Go code. The generated function will automatically unzip
	// the file data when called. Defaults to false.
	NoCompress bool

	// Perform a debug build. This generates an asset file, which
	// loads the asset contents directly from disk at their original
	// location, instead of embedding the contents in the code.
	//
	// This is mostly useful if you anticipate that the assets are
	// going to change during your development cycle. You will always
	// want your code to access the latest version of the asset.
	// Only in release mode, will the assets actually be embedded
	// in the code. The default behaviour is Release mode.
	Debug bool

	// Perform a dev build, which is nearly identical to the debug option. The
	// only difference is that instead of absolute file paths in generated code,
	// it expects a variable, `rootDir`, to be set in the generated code's
	// package (the author needs to do this manually), which it then prepends to
	// an asset's name to construct the file path on disk.
	//
	// This is mainly so you can push the generated code file to a shared
	// repository.
	Dev bool

	// When true, size, mode and modtime are not preserved from files
	NoMetadata bool
	// When nonzero, use this as mode for all files.
	Mode uint
	// When nonzero, use this as unix timestamp for all files.
	ModTime int64
	// When nonzero, use this as unix timestamp for all files.
	ChangeTime int64

	// Ignores any filenames matching the regex pattern specified, e.g.
	// path/to/file.ext will ignore only that file, or \\.gitignore
	// will match any .gitignore file.
	//
	// This parameter can be provided multiple times.
	Ignore []*regexp.Regexp

	// Ignores any filenames matching the glob pattern specified, e.g.
	// path/to/*.ext will ignore only that file, or \\.gitignore
	// will match any .gitignore file.
	// See <https://github.com/gobwas/glob> for more glob info.
	//
	// This parameter can be provided multiple times.
	IgnoreGlob []glob.Glob

	// Create File System.
	//
	// This parameter provides `AssetFS` variable.
	FileSystem bool

	// Outlined assets content into binary file. Do not compile it.
	Outlined bool

	OutlinedOutputDir string

	OutlinedLocalOutputDir string

	//
	OutlineEmbeded bool

	// OutlinedApi save asset contents into him. If is blank, use the application binary.
	OutlinedApi string

	// OutlinedNoTruncate truncate existing outlined.
	OutlinedNoTruncate bool

	OutlinedProgram bool

	EmbedPreInitSource string

	OutlinedHeadersOutput string

	NoAutoLoad bool

	Hybrid bool

	NoStore bool

	OulinedSkipApi bool

	InputProduction bool

	FileSystemLoadCallbacks []string
}

// NewConfig returns a default configuration struct.
func NewConfig() *Config {
	c := new(Config)
	c.Package = "main"
	c.NoMemCopy = false
	c.NoCompress = false
	c.Debug = false
	c.Output = DefaultOutput
	return c
}

// validate ensures the config has sane values.
// Part of which means checking if certain file/directory paths exist.
func (c *Config) validate() error {
	if c.Package == "" {
		return fmt.Errorf("Missing package name")
	}

	for _, input := range c.Input {
		_, err := os.Lstat(input.Path)
		if err != nil {
			return fmt.Errorf("Failed to stat input path '%s': %v", input.Path, err)
		}
	}

	if c.Outlined {
		if c.OutlinedOutputDir == "" {
			c.OutlinedOutputDir = "_assets"
		}

		if c.OutlinedLocalOutputDir == "" {
			c.OutlinedLocalOutputDir = filepath.Base(c.OutlinedOutputDir)
		}

		if c.Output == "" {
			if c.OutlinedProgram {
				c.Output = OutputToProgram
			} else {
				c.Output = c.Package + ".xb"
			}
		}
	} else if c.Output == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("Unable to determine current working directory.")
		}

		c.Output = filepath.Join(cwd, "xb.go")
	}

	stat, err := os.Lstat(c.Output)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Output path: %v", err)
		}

		// File does not exist. This is fine, just make
		// sure the directory it is to be in exists.
		dir, _ := filepath.Split(c.Output)
		if dir != "" {
			err = os.MkdirAll(dir, 0744)

			if err != nil {
				return fmt.Errorf("Create output directory: %v", err)
			}
		}
	}

	if stat != nil && stat.IsDir() {
		return fmt.Errorf("Output path is a directory.")
	}

	if c.Hybrid {
		c.FileSystem = true
	}

	return nil
}
