package www

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type Server interface {
	RenderTemplate(w http.ResponseWriter, name string, data interface{})
}

type server struct {
	templates map[string]*template.Template
}

func Init(funcs template.FuncMap) Server {
	s := &server{
		templates: make(map[string]*template.Template),
	}

	const mainTmpl = `{{ define "main" }} {{ template "base" . }} {{ end }}`

	layoutFiles, err := filepath.Glob("www/templates/layouts/*.gohtml")
	if err != nil {
		log.Fatal(err)
	}

	includeFiles, err := filepath.Glob("www/templates/*.gohtml")
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
		s.templates[fileName], err = mainTemplate.Clone()
		if err != nil {
			log.Fatal(err)
		}
		s.templates[fileName] = template.Must(s.templates[fileName].Funcs(funcs).ParseFiles(files...))
	}

	return s
}

func (s *server) RenderTemplate(w http.ResponseWriter, name string, data interface{}) {
	tmpl, ok := s.templates[name]
	if !ok {
		http.Error(w, fmt.Sprintf("The template %s does not exist.", name),
			http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
