// Code generated for package api by go-bindata DO NOT EDIT. (@generated)
// sources:
// webui/android-chrome-192x192.png
// webui/android-chrome-512x512.png
// webui/apple-touch-icon.png
// webui/css/bulma.css
// webui/css/bulma.css.map
// webui/css/bulma.min.css
// webui/css/loader.css
// webui/favicon-16x16.png
// webui/favicon-32x32.png
// webui/favicon.ico
// webui/index.html
// webui/js/FormData.js
// webui/js/actions.js
// webui/js/app.js
// webui/js/fetch.js
// webui/js/fontawesome.js
// webui/js/kala.js
// webui/js/promise.js
// webui/js/reef/reef.polyfills.min.js
// webui/js/reef/router.min.js
// webui/js/routes.js
// webui/js/store.js
// webui/js/utils.js
// webui/logo.png
// webui/site.webmanifest
package api

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	assetfs "github.com/elazarl/go-bindata-assetfs"
)

func bindataRead(name string) (*webuiData, error) {
	data, ok := webuiDataMap[name]
	if !ok {
		return nil, fmt.Errorf("Read %q", name)
	}
	gz, err := gzip.NewReader(bytes.NewBuffer(data.body))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}
	data.body = buf.Bytes()
	return &data, nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

func assetFunc(name string) (*asset, error) {
	data, err := bindataRead(name)
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: name, size: data.fileInfo.Size(), mode: os.FileMode(420), modTime: data.fileInfo.ModTime()}
	a := &asset{bytes: data.body, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if _bindata[cannonicalName] {
		a, err := assetFunc(cannonicalName)
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if _bindata[cannonicalName] {
		a, err := assetFunc(cannonicalName)
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]bool{
	"webui/android-chrome-192x192.png":    true,
	"webui/android-chrome-512x512.png":    true,
	"webui/apple-touch-icon.png":          true,
	"webui/css/bulma.css":                 true,
	"webui/css/bulma.css.map":             true,
	"webui/css/bulma.min.css":             true,
	"webui/css/loader.css":                true,
	"webui/favicon-16x16.png":             true,
	"webui/favicon-32x32.png":             true,
	"webui/favicon.ico":                   true,
	"webui/index.html":                    true,
	"webui/js/FormData.js":                true,
	"webui/js/actions.js":                 true,
	"webui/js/app.js":                     true,
	"webui/js/fetch.js":                   true,
	"webui/js/fontawesome.js":             true,
	"webui/js/kala.js":                    true,
	"webui/js/promise.js":                 true,
	"webui/js/reef/reef.polyfills.min.js": true,
	"webui/js/reef/router.min.js":         true,
	"webui/js/routes.js":                  true,
	"webui/js/store.js":                   true,
	"webui/js/utils.js":                   true,
	"webui/logo.png":                      true,
	"webui/site.webmanifest":              true,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func(string) (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"webui": &bintree{nil, map[string]*bintree{
		"android-chrome-192x192.png": &bintree{assetFunc, map[string]*bintree{}},
		"android-chrome-512x512.png": &bintree{assetFunc, map[string]*bintree{}},
		"apple-touch-icon.png":       &bintree{assetFunc, map[string]*bintree{}},
		"css": &bintree{nil, map[string]*bintree{
			"bulma.css":     &bintree{assetFunc, map[string]*bintree{}},
			"bulma.css.map": &bintree{assetFunc, map[string]*bintree{}},
			"bulma.min.css": &bintree{assetFunc, map[string]*bintree{}},
			"loader.css":    &bintree{assetFunc, map[string]*bintree{}},
		}},
		"favicon-16x16.png": &bintree{assetFunc, map[string]*bintree{}},
		"favicon-32x32.png": &bintree{assetFunc, map[string]*bintree{}},
		"favicon.ico":       &bintree{assetFunc, map[string]*bintree{}},
		"index.html":        &bintree{assetFunc, map[string]*bintree{}},
		"js": &bintree{nil, map[string]*bintree{
			"FormData.js":    &bintree{assetFunc, map[string]*bintree{}},
			"actions.js":     &bintree{assetFunc, map[string]*bintree{}},
			"app.js":         &bintree{assetFunc, map[string]*bintree{}},
			"fetch.js":       &bintree{assetFunc, map[string]*bintree{}},
			"fontawesome.js": &bintree{assetFunc, map[string]*bintree{}},
			"kala.js":        &bintree{assetFunc, map[string]*bintree{}},
			"promise.js":     &bintree{assetFunc, map[string]*bintree{}},
			"reef": &bintree{nil, map[string]*bintree{
				"reef.polyfills.min.js": &bintree{assetFunc, map[string]*bintree{}},
				"router.min.js":         &bintree{assetFunc, map[string]*bintree{}},
			}},
			"routes.js": &bintree{assetFunc, map[string]*bintree{}},
			"store.js":  &bintree{assetFunc, map[string]*bintree{}},
			"utils.js":  &bintree{assetFunc, map[string]*bintree{}},
		}},
		"logo.png":         &bintree{assetFunc, map[string]*bintree{}},
		"site.webmanifest": &bintree{assetFunc, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

func AssetFS() *assetfs.AssetFS {
	assetInfo := func(path string) (os.FileInfo, error) {
		return os.Stat(path)
	}
	for k := range _bintree.Children {
		return &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: assetInfo, Prefix: k}
	}
	panic("unreachable")
}
