package main

import (
	"encoding/json"
	"flag"
	"github.com/catcombo/go-staticfiles"
	"github.com/vsliouniaev/go-pass-cache/cache"
	"github.com/vsliouniaev/go-pass-cache/util"
	"github.com/vsliouniaev/go-pass-cache/www"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	c               cache.Cache
	ww              www.Server
	maxSize         int64
	linkProbeAgents = map[string]struct{}{
		"skype": {}, "whatsapp": {}, "slack": {}, "signal": {}, "telegram": {}, "zoom": {},
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
		ww.RenderTemplate(w, "set.gohtml", nil)
	} else {
		val, ok := c.TryGet(key)
		if !ok {
			ww.RenderTemplate(w, "gone.gohtml", nil)
		} else {
			ww.RenderTemplate(w, "get.gohtml", val)
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

// TODO: Propagate cache duration to UI
// TODO: Remove

func main() {
	var (
		bind          string
		dev           bool
		cacheDuration time.Duration
		ignore        util.ArrayFlags
	)
	flag.BoolVar(&dev, "dev", false, "development mode")
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

	staticLoc := "/www/static/"
	funcs, staticHandler := initStatic(staticLoc, dev)
	ww = www.Init(funcs)
	s := http.Server{Addr: bind}
	if dev {
		http.HandleFunc(staticLoc, util.NoCache(util.WithGzip(staticHandler)))
	} else {
		http.HandleFunc(staticLoc, util.CacheForever(util.WithGzip(staticHandler)))
	}

	http.HandleFunc("/", util.NoCache(util.WithGzip(genericWithFilter)))
	log.Println(s.ListenAndServe())
}

// TODO: There is a discrepancy between passing paths to this and to the template server
//  but it's awkward because of the slashes that have to be added to everything
func initStatic(staticLoc string, dev bool) (template.FuncMap, http.HandlerFunc) {
	storage, err := staticfiles.NewStorage(".static")
	if err != nil {
		log.Fatal(err)
	}
	storage.AddInputDir(strings.Trim(staticLoc, "/"))
	err = storage.CollectStatic()
	if err != nil {
		log.Fatal(err)
	}

	funcs := template.FuncMap{
		"static": func(relPath string) string {
			return staticLoc + storage.Resolve(relPath)
		},
	}

	storage.OutputDirList = false
	storage.Enabled = !dev

	return funcs, http.StripPrefix(staticLoc, http.FileServer(storage)).ServeHTTP
}
