package main

import (
	"context"
	"flag"
	"github.com/hneemann/control/server"
	"github.com/hneemann/control/server/data"
	"github.com/hneemann/session"
	"github.com/hneemann/session/fileSys"
	"github.com/hneemann/session/myOidc"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

type persist struct{}

func retry(count int, delay time.Duration, task func() bool) bool {
	for i := range count {
		if i > 0 {
			log.Printf("Retry %d/%d in %v", i, count, delay)
			time.Sleep(delay)
		}
		if task() {
			return true
		}
	}
	return false
}

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
	debug := flag.Bool("debug", false, "debug mode. In this mode, the server does not enable browser caching. Also, user 'admin' with password 'admin' is created with a fixed session token. This does not work if OIDC is used!")
	onServer := flag.Bool("onServer", false, "execution on server, otherwise in browser")
	oidc := flag.Bool("oidc", false, "oidc mode")
	flag.Parse()

	log.Println("Folder:", *dataFolder)

	var sc *session.Cache[data.UserData]
	if *oidc {
		sc = session.NewSessionCache[data.UserData](
			myOidc.NewOidcDataManager[data.UserData](
				session.NewFileSystemFactory(*dataFolder),
				persist{}),
			4*time.Hour, time.Hour)
	} else {
		sc = session.NewSessionCache[data.UserData](
			session.NewDataManager[data.UserData](
				session.NewFileSystemFactory(*dataFolder),
				persist{}),
			4*time.Hour, time.Hour)
		if *debug {
			err := sc.CreateDebugSession("admin", "admin", "debugTokenForAdmin")
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	defer sc.Close()

	mux := http.NewServeMux()

	if *oidc {
		isOidc := retry(3, time.Second*3, func() bool {
			return myOidc.RegisterLogin(mux, "/login", "/auth/callback", myOidc.CreateOidcSession(sc))
		})
		if !isOidc {
			log.Fatal("oidc not available!")
		}
	} else {
		mux.HandleFunc("/login", sc.LoginHandler(server.Templates.Lookup("login.html")))
		mux.HandleFunc("/register", sc.RegisterHandler(server.Templates.Lookup("register.html")))
	}

	examples := server.ReadExamples()
	mux.HandleFunc("/", sc.CheckSessionFunc(server.CreateMain(examples)))
	mux.Handle("/assets/", Cache(http.FileServer(http.FS(server.Assets)), 180, !*debug))
	mux.Handle("/js/execute.js", Cache(server.RunMode(*onServer), 180, !*debug))
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
