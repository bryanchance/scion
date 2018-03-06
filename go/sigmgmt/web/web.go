// Copyright 2018 Anapaya Systems

package web

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
	texttemplate "text/template"
	"time"

	log "github.com/inconshreveable/log15"

	"github.com/scionproto/scion/go/sigmgmt/config"
	"github.com/scionproto/scion/go/sigmgmt/db"
	"github.com/scionproto/scion/go/sigmgmt/web/wikidoc"
	"github.com/scionproto/scion/go/sigmgmt/weblogger"
)

const (
	// Total timeout for pushes (config generation, saving, copying, reloading,
	// verification)
	DefaultTimeout = 5 * time.Second
)

// Page contains information for the top-level frontend page
type Page struct {
	Body     []byte
	Features config.FeatureLevel
}

type Application struct {
	cfg              *config.Global
	dbase            *db.DB
	mux              *http.ServeMux
	wikiServer       *wikidoc.WikiServer
	asFormTemplate   *template.Template
	siteFormTemplate *template.Template
	// HTML sanitization happens during page content rendering in their
	// respective templates, so no sanitization is needed for top level page.
	topTemplate *texttemplate.Template
}

func NewApplication(cfg *config.Global, dbase *db.DB) *Application {
	asFuncMap := map[string]interface{}{
		"getPolicyFromIA":         getPolicyFromIA,
		"getSessionAliasesFromIA": getSessionAliasesFromIA,
		"sortByIA":                sortByIA,
		"iaIntToIA":               iaIntToIA,
		"pairs":                   weblogger.Pairs,
		"upper":                   weblogger.PrintLogLevel,
	}

	siteFuncMap := map[string]interface{}{
		"pairs": weblogger.Pairs,
		"upper": weblogger.PrintLogLevel,
	}

	a := &Application{
		cfg:        cfg,
		dbase:      dbase,
		mux:        http.NewServeMux(),
		wikiServer: &wikidoc.WikiServer{},
		asFormTemplate: template.Must(
			template.New("as.html").Funcs(asFuncMap).ParseFiles(
				filepath.Join(cfg.WebAssetRoot, "as.html"),
				filepath.Join(cfg.WebAssetRoot, "weblogger/logs.html"),
			),
		),
		siteFormTemplate: template.Must(
			template.New("sites.html").Funcs(siteFuncMap).ParseFiles(
				filepath.Join(cfg.WebAssetRoot, "sites.html"),
				filepath.Join(cfg.WebAssetRoot, "weblogger/logs.html"),
			),
		),
		topTemplate: texttemplate.Must(
			texttemplate.ParseFiles(
				filepath.Join(cfg.WebAssetRoot, "page.html"),
			),
		),
	}
	a.mux.Handle("/static/", http.FileServer(http.Dir(".")))
	a.mux.HandleFunc("/static/doc/", a.HandleDoc)
	a.mux.HandleFunc("/config/as", a.HandleASConfig)
	a.mux.HandleFunc("/config/sites", a.HandleSiteConfig)
	a.mux.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/config/sites", http.StatusFound)
		})
	return a
}

func (a *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func (a *Application) HandleDoc(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path[0] == '/' {
		path = path[1:]
	}
	body, err := a.wikiServer.Load(path)
	if err != nil {
		log.Error("wiki error", "err", err)
		http.Error(w, http.StatusText(404), 404)
	} else {
		a.ServePage(w, Page{
			Body: body,
		})
	}
}

func (a *Application) HandleASConfig(w http.ResponseWriter, r *http.Request) {
	data := BuildASFormData(a.cfg, a.dbase, r)
	body := new(bytes.Buffer)
	if err := a.asFormTemplate.Execute(body, data); err != nil {
		log.Error("template error", "err", err)
		http.Error(w, err.Error(), 500)
	} else {
		a.ServePage(w, Page{
			Body:     body.Bytes(),
			Features: a.cfg.Features,
		})
	}
}

func (a *Application) HandleSiteConfig(w http.ResponseWriter, r *http.Request) {
	data := BuildSiteFormData(a.cfg, a.dbase, r)
	body := new(bytes.Buffer)
	if err := a.siteFormTemplate.Execute(body, data); err != nil {
		log.Error("template error", "err", err)
		http.Error(w, err.Error(), 500)
	} else {
		a.ServePage(w, Page{
			Body:     body.Bytes(),
			Features: a.cfg.Features,
		})
	}
}

func (a *Application) ServePage(w http.ResponseWriter, page Page) {
	if err := a.topTemplate.ExecuteTemplate(w, "page.html", page); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
