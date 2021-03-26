package main

import (
	"encoding/json"
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

type Data struct {
	Id   string `json:"id"`
	Data string `json:"data"`
}

func generic(w http.ResponseWriter, r *http.Request) {
	key := r.URL.RawQuery
	if key == "" {
		if r.ContentLength < maxSize*1000000 {
			var data Data
			if err := json.NewDecoder(r.Body).Decode(&data); err == nil {
				if data.Id != "" && data.Data != "" {
					c.AddKey(data.Id, data.Data)
				}
			}
		}
		renderTemplate(w, "set.gohtml", nil)
	} else {
		val, ok := c.TryGet(key)
		if !ok {
			renderTemplate(w, "gone.gohtml", nil)
		} else {
			renderTemplate(w, "get.gohtml", val)
		}
	}
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

func genericWithFilter(w http.ResponseWriter, r *http.Request) {
	if filterLinkProbes(w, r) {
		generic(w, r)
	}
}

func main() {
	var (
		bind          string
		cacheDuration time.Duration
		ignore        arrayFlags
	)
	flag.StringVar(&bind, "bind", ":8080", "address:port to bind to. Default :8080")
	flag.Var(&ignore, "ignore-agents", "Ignore user-agent strings containing this value. Flag can be specified multiple times.")
	flag.DurationVar(&cacheDuration, "cache-duration", time.Minute*5, "Cache duration. Default 5m")
	flag.Int64Var(&maxSize, "max-size", 10, "Max size of request in MB. Default 10MB")

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
	http.HandleFunc("/", genericWithFilter)
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
