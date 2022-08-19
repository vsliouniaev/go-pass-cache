package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	gz "github.com/NYTimes/gziphandler"
	"github.com/catcombo/go-staticfiles"

	"github.com/vsliouniaev/go-pass-cache/cache"
	"github.com/vsliouniaev/go-pass-cache/util"
	"github.com/vsliouniaev/go-pass-cache/www"
)

var (
	c               cache.Cache
	ww              www.Server
	maxSize         int64
	cacheDuration   time.Duration
	linkProbeAgents = map[string]struct{}{
		"discord": {}, "skype": {}, "whatsapp": {}, "slack": {}, "signal": {}, "telegram": {}, "zoom": {},
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
					c.Store(data.Id, data.Data)
				}
			}
		}
		ww.RenderTemplate(w, "set.gohtml", cacheDuration)
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

type genericWithFilter struct{}

func (g *genericWithFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if filterLinkProbes(w, r) {
		generic(w, r)
	}
}

func main() {
	var (
		bind   string
		dev    bool
		ignore util.ArrayFlags
	)
	flag.BoolVar(&dev, "dev", false, "Development mode. Default false")
	flag.StringVar(&bind, "bind", ":8080", "address:port to bind to. Default :8080")
	flag.Var(&ignore, "ignore-agents",
		fmt.Sprintf("Ignore user-agent strings containing this value. Flag can be specified multiple times. Default %s)",
			strings.Join(util.SortedKeys(linkProbeAgents), ", ")))
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
	fns, staticHandler := initStatic(staticLoc, dev)
	ww = www.Init(fns)
	var strat util.CacheStrategy
	if dev {
		strat = util.Never
	} else {
		strat = util.Forever
	}
	http.Handle(staticLoc, util.NewCacheHandler(strat, gz.GzipHandler(staticHandler)))
	http.Handle("/", util.NewCacheHandler(util.Never, gz.GzipHandler(&genericWithFilter{})))

	if err := http.ListenAndServe(bind, nil); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func initStatic(staticLoc string, dev bool) (template.FuncMap, http.Handler) {
	storage, err := staticfiles.NewStorage(".static")
	if err != nil {
		log.Fatal(err)
	}
	storage.AddInputDir(strings.Trim(staticLoc, "/"))
	if err = storage.CollectStatic(); err != nil {
		log.Fatal(err)
	}

	fns := template.FuncMap{
		"static": func(relPath string) string {
			return staticLoc + storage.Resolve(relPath)
		},
	}

	storage.OutputDirList = false
	storage.Enabled = !dev

	return fns, http.StripPrefix(staticLoc, http.FileServer(storage))
}
