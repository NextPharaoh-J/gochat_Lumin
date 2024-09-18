package site

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gochat_my/config"
	"net/http"
	"os"
	"path"
)

type Site struct{}

func New() *Site {
	return &Site{}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	// 404 back
	data, _ := os.ReadFile("./site/index.html")
	_, _ = fmt.Fprintf(w, string(data))
	return
}

func server(fs http.FileSystem) http.Handler {
	fileServer := http.FileServer(fs)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filePath := path.Clean(r.URL.Path) //legalize file path
		_, err := os.Stat(filePath)
		if err != nil {
			notFound(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

func (s *Site) Run() {
	port := config.Conf.Site.SiteBase.ListenPort
	addr := fmt.Sprintf(":%d", port)
	logrus.Fatal(http.ListenAndServe(addr, server(http.Dir("./site"))))
}
