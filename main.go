package main

import (
	"context"
	"flag"
	"github.com/hneemann/control/server"
	"github.com/hneemann/control/server/data"
	"github.com/hneemann/session"
	"github.com/hneemann/session/fileSys"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

type persist struct{}

func (p persist) Load(f fileSys.FileSystem) (*data.UserData, error) {
	r, err := f.Reader("data.json")
	if err != nil {
		return nil, err
	}
	defer fileSys.CloseLog(r)
	return data.Load(r)
}

func (p persist) Save(f fileSys.FileSystem, data *data.UserData) error {
	w, err := f.Writer("data.json")
	if err != nil {
		return err
	}
	defer fileSys.CloseLog(w)
	return data.Save(w)
}

func main() {
	dataFolder := flag.String("folder", "", "data folder")
	cert := flag.String("cert", "", "certificate")
	key := flag.String("key", "", "certificate")
	port := flag.Int("port", 8080, "port")
	debug := flag.Bool("debug", false, "debug mode")
	flag.Parse()

	log.Println("Folder:", *dataFolder)
	sc := session.NewSessionCache[data.UserData](
		session.NewDataManager[data.UserData](
			session.NewFileSystemFactory(*dataFolder),
			persist{}),
		4*time.Hour, time.Hour)
	defer sc.Close()

	examples := server.ReadExamples()

	mux := http.NewServeMux()
	if *debug {
		mux.HandleFunc("/", sc.DebugLogin("admin", "admin", server.CreateMain(examples)))
	} else {
		mux.HandleFunc("/", sc.CheckSessionFunc(server.CreateMain(examples)))
	}
	mux.HandleFunc("/login", sc.LoginHandler(server.Templates.Lookup("login.html")))
	mux.HandleFunc("/register", sc.RegisterHandler(server.Templates.Lookup("register.html")))

	mux.Handle("/assets/", Cache(http.FileServer(http.FS(server.Assets)), 60, !*debug))
	mux.HandleFunc("/execute/", sc.CheckSessionRestFunc(server.Execute))
	mux.HandleFunc("/example/", sc.CheckSessionRestFunc(server.CreateExamples(examples)))
	mux.HandleFunc("/files/", sc.CheckSessionRestFunc(server.Files))

	serv := &http.Server{Addr: ":" + strconv.Itoa(*port), Handler: mux}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Print("interrupted")

		err := serv.Shutdown(context.Background())
		if err != nil {
			log.Println(err)
		}
		for {
			<-c
		}
	}()

	var err error
	if *cert != "" && *key != "" {
		log.Println("Starting server with TLS")
		err = serv.ListenAndServeTLS(*cert, *key)
	} else {
		log.Println("Starting server without TLS")
		err = serv.ListenAndServe()
	}
	if err != nil {
		log.Println(err)
	}

}

func Cache(parent http.Handler, minutes int, enableCache bool) http.Handler {
	if enableCache {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Add("Cache-Control", "public, max-age="+strconv.Itoa(minutes*60))
			parent.ServeHTTP(writer, request)
		})
	} else {
		log.Println("browser caching disabled")
		return parent
	}
}
