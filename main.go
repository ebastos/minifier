package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/juju/loggo"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"
)

var logger = loggo.GetLogger("main")
var rootLogger = loggo.GetLogger("")

var (
	port     = flag.String("port", ":3000", "Port on which it will listen.")
	logLevel = flag.String("logLevel", "INFO", "TRACE, DEBUG, INFO, WARNING, ERROR, CRITICAL")
)

func init() {
	flag.Parse()
	loggo.ConfigureLoggers(*logLevel)
	rootLogger.Infof("Starting...")

}

func return500(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 - Something bad happened!"))
}

func minifier(f string, b []byte) (payload []byte, ftype string, e error) {
	// Let's start saying it's a text/html and we will minify
	ctype := "text/html"
	var err error

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)

	switch mtype := strings.ToLower(filepath.Ext(f)); mtype {
	case ".css":
		b, err = m.Bytes("text/css", b)
		if err != nil {
			rootLogger.Errorf("Something went wrong %s", err)
			return b, ctype, err
		}
		ctype = "text/css"
	case ".js":
		b, err = m.Bytes("application/javascript", b)
		if err != nil {
			rootLogger.Errorf("Something went wrong %s", err)
			return b, ctype, err
		}
		ctype = "application/javascript"
	default:
		// How the heck we ended up with an unknown extension?
		rootLogger.Warningf("This should not happen. Is Nginx misconfigured? %s", f)

	}
	return b, ctype, nil
}

func reqHandler(w http.ResponseWriter, r *http.Request) {
	rootLogger.Debugf("Invoked for %s", r.URL.Path)

	minime := "true"
	ctype := "text/html"
	var err error

	headers := w.Header()
	// Physical file path has to be sent by Nginx via X-Docroot header.
	if len(r.Header["X-Docroot"]) != 1 {
		rootLogger.Errorf("Missing X-Docroot header from proxy! Refusing to run")
		return500(w)
		return
	}
	mediaFile := r.URL.Path
	filePath := r.Header["X-Docroot"][0] + mediaFile

	// Let's try to open the file. 404 if not there.
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		rootLogger.Errorf("failed %s", err)
		http.NotFound(w, r)
	}

	//If there is a query argument ?minify=false we just serve the original file.
	if len(r.URL.Query()["minify"]) == 1 {
		minime = r.URL.Query()["minify"][0]
	}
	if minime != "false" {
		// If not let's minify it.
		b, ctype, err = minifier(mediaFile, b)
		if err != nil {
			rootLogger.Errorf("failed minifying: %s", err)
			return500(w)
		}
		headers.Add("Content-Type", ctype)
	}
	io.WriteString(w, string(b))
}

func main() {
	http.HandleFunc("/", reqHandler)
	log.Fatal(http.ListenAndServe(*port, nil))
}
