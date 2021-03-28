package www

import (
	"fmt"
	"github.com/catcombo/go-staticfiles"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type Server interface {
	http.Handler
	RenderTemplate(w http.ResponseWriter, name string, data interface{})
}

type server struct {
	templates map[string]*template.Template
	mux       *http.ServeMux
}

func Init(templatePath, staticPath string) Server {
	s := &server{
		templates: make(map[string]*template.Template),
	}

	const mainTmpl = `{{ define "main" }} {{ template "base" . }} {{ end }}`

	layoutFiles, err := filepath.Glob(templatePath + "/layouts/*.gohtml")
	if err != nil {
		log.Fatal(err)
	}

	includeFiles, err := filepath.Glob(templatePath + "/*.gohtml")
	if err != nil {
		log.Fatal(err)
	}

	mainTemplate := template.New("main")

	mainTemplate, err = mainTemplate.Parse(mainTmpl)
	if err != nil {
		log.Fatal(err)
	}

	staticFilesPrefix := "/" + staticPath + "/"
	staticFilesRoot := ".static"

	storage, err := staticfiles.NewStorage(staticFilesRoot)
	if err != nil {
		log.Fatal(err)
	}
	storage.AddInputDir(staticPath)
	err = storage.CollectStatic()
	if err != nil {
		log.Fatal(err)
	}

	funcs := template.FuncMap{
		"static": func(relPath string) string {
			return staticFilesPrefix + storage.Resolve(relPath)
		},
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

	storage.OutputDirList = false
	handler := http.StripPrefix(staticFilesPrefix, http.FileServer(storage))
	s.mux = http.NewServeMux()
	s.mux.Handle(staticFilesPrefix, handler)
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

func (s *server) ServeHTTP(r http.ResponseWriter, w *http.Request) {
	s.mux.ServeHTTP(r, w)
}
