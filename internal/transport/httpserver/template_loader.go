package httpserver

import (
	"html/template"
	"io/fs"
	"sort"
	"strings"
)

func mustParseTemplates(fsys fs.FS) *template.Template {
	var files []string
	err := fs.WalkDir(fsys, "template", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".html") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}

	sort.Strings(files)
	return template.Must(template.New("").ParseFS(fsys, files...))
}
