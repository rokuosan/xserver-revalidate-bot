package main

import (
	"errors"
	"html/template"
	"io"
)

type TemplateRenderer struct {
	template *template.Template
}

func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{}
}

func (tr *TemplateRenderer) LoadTemplate(templatePath string) error {
	var err error
	tr.template, err = template.ParseFiles(templatePath)
	return err
}

func (tr *TemplateRenderer) Render(w io.Writer, data TemplateData) error {
	if tr.template == nil {
		return errors.New("template not loaded")
	}

	return tr.template.Execute(w, data)
}
