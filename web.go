package main

import (
	"fmt"
	"github.com/oxtoacart/bpool"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var templates map[string]*template.Template
var bufpool *bpool.BufferPool

type TemplateConfig struct {
	TemplateLayoutPath  string
	TemplateIncludePath string
}

var mainTmpl = `{{define "main" }} {{ template "base" . }} {{ end }}`

var templateConfig TemplateConfig

func loadConfiguration() {
	templateConfig.TemplateLayoutPath = "www/templates/layouts/"
	templateConfig.TemplateIncludePath = "www/templates/"
}

func loadTemplates() {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}


	layoutFiles, err := filepath.Glob(templateConfig.TemplateLayoutPath + "*.gohtml")
	if err != nil {
		log.Fatal(err)
	}

	includeFiles, err := filepath.Glob(templateConfig.TemplateIncludePath + "*.gohtml")
	if err != nil {
		log.Fatal(err)
	}

	mainTemplate := template.New("main")

	mainTemplate, err = mainTemplate.Parse(mainTmpl)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range includeFiles {
		fileName := filepath.Base(file)
		files := append(layoutFiles, file)
		templates[fileName], err = mainTemplate.Clone()
		if err != nil {
			log.Fatal(err)
		}
		templates[fileName] = template.Must(templates[fileName].ParseFiles(files...))
	}

	log.Printf("templates loading successful %d\n", len(templates))

	bufpool = bpool.NewBufferPool(64)
	log.Println("buffer allocation successful")
}

func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	tmpl, ok := templates[name]
	if !ok {
		http.Error(w, fmt.Sprintf("The template %s does not exist.", name),
			http.StatusInternalServerError)
	}

	buf := bufpool.Get()
	defer bufpool.Put(buf)

	err := tmpl.Execute(buf, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}