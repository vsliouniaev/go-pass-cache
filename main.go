package main

import (
	"log"
	"net/http"
	"strings"
	"time"
)

var cache = Cache{
	data: map[string]CacheObject{},
}

func set(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("id")
	val := r.Form.Get("data")

	if key != "" && val != "" {
		cache.AddOrSilentlyFail(key, val)
	}
	renderTemplate(w, "set.gohtml", getEntropy())
}

func get(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("id")
	if key == "" {
		renderTemplate(w, "gone.gohtml", nil)
	}

	val, ok := cache.TryGetAndRemoveWithinTimeFrame(key, time.Minute*5);
	if !ok {
		renderTemplate(w, "gone.gohtml", nil)
	}

	renderTemplate(w, "get.gohtml", val)
}

func redirectRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/set", http.StatusSeeOther)
}

func filterLinkProbes(w http.ResponseWriter, r *http.Request) bool {
	ua := r.Header.Get("User-Agent")

	if
	strings.Contains(ua, "skype") ||
		strings.Contains(ua, "whatsapp", ) ||
		strings.Contains(ua, "slack") {
		w.WriteHeader(404)
		return false
	}

	return true
}

func setWithFilter(w http.ResponseWriter, r *http.Request) {
	if filterLinkProbes(w, r) {
		set(w, r)
	}
}

func getWithFilter(w http.ResponseWriter, r *http.Request) {
	if filterLinkProbes(w, r) {
		get(w, r)
	}
}

func main() {
	loadConfiguration()
	loadTemplates()
	s := http.Server{Addr: ":8080"}
	http.HandleFunc("/set", setWithFilter)
	http.HandleFunc("/get", getWithFilter)
	fs := http.FileServer(http.Dir("www/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", redirectRoot)
	log.Println(s.ListenAndServe())
}
