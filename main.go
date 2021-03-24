package main

import (
	"flag"
	"fmt"
	"github.com/vsliouniaev/go-pass-cache/cache"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	c               cache.Cache
	maxSize         int64
	linkProbeAgents = map[string]struct{}{
		"skype": {}, "whatsapp": {}, "slack": {},
	}
)

func set(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength < maxSize*1000000 {
		r.ParseForm()
		key := r.Form.Get("id")
		val := r.Form.Get("data")

		if key != "" && val != "" {
			c.AddKey(key, val)
		}
	}
	renderTemplate(w, "set.gohtml", getEntropy())
}

func get(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("id")
	if key == "" {
		renderTemplate(w, "gone.gohtml", nil)
	}

	val, ok := c.TryGet(key)
	if !ok {
		renderTemplate(w, "gone.gohtml", nil)
	}

	renderTemplate(w, "get.gohtml", val)
}

func filterLinkProbes(w http.ResponseWriter, r *http.Request) bool {
	ua := r.Header.Get("User-Agent")

	for a := range linkProbeAgents {
		if strings.Contains(ua, a) {
			w.WriteHeader(http.StatusNoContent)
			return false
		}
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
	var (
		bind          string
		cacheDuration time.Duration
		ignore        arrayFlags
	)
	flag.Int64Var(&maxSize, "max-size", 10, "Max size of request in MB. Default 10MB")
	flag.DurationVar(&cacheDuration, "cache-duration", time.Minute*5, "Cache duration. Default 5m")
	flag.StringVar(&bind, "bind", ":8080", "address:port to bind to. Default :8080")
	flag.Var(&ignore, "ignore-agents", "Ignore user-agent strings containing this value. Flag can be specified multiple times.")
	flag.Parse()

	if maxSize < 0 {
		log.Println("max-size cannot be negative")
	}
	if cacheDuration < time.Second*5 {
		log.Println("cache-duration < 5s is ridiculous")
	}
	for _, strMatch := range ignore {
		linkProbeAgents[strings.ToLower(strMatch)] = struct{}{}
	}

	c = cache.New(cacheDuration)

	loadTemplates()
	s := http.Server{Addr: bind}
	fs := http.FileServer(http.Dir("www/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/get", getWithFilter)
	http.HandleFunc("/", setWithFilter)
	log.Println(s.ListenAndServe())
}

type arrayFlags []string

func (af *arrayFlags) String() string {
	return fmt.Sprintf("%s", []string(*af))
}

func (af *arrayFlags) Set(value string) error {
	*af = append(*af, value)
	return nil
}
