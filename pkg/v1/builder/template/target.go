package template

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

// Target to template files.
type Target struct {
	// Path of the file.
	Filepath string

	// Template to use.
	Template string
}

// Run the target.
func (t Target) Run(data interface{}) error {
	tmplFp := template.Must(template.New("target-fp").Parse(t.Filepath))
	bufFp := &bytes.Buffer{}
	err := tmplFp.Execute(bufFp, data)
	if err != nil {
		return err
	}
	fp := bufFp.String()

	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("target").Parse(t.Template))
	if err := tmpl.Execute(buf, data); err != nil {
		return err
	}

	dir := filepath.Dir(fp)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	err = ioutil.WriteFile(fp, buf.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}
