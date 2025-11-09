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
	"syscall"
	"time"
)

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

type persist struct{}

func (p persist) Init(_ fileSys.FileSystem, _ *data.UserData) error {
	return nil
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

	var dm session.Manager[data.UserData]
	dm = session.NewFileManager[data.UserData](
		session.NewFileSystemFactory(*dataFolder),
		persist{})

	if *oidc {
		dm = myOidc.NewOidcDataManager[data.UserData](dm)
	}
	sc := session.NewSessionCache[data.UserData](dm, 4*time.Hour, time.Hour)
	if *debug && !*oidc {
		err := sc.CreateDebugSession("admin", "admin", "debugTokenForAdmin")
		if err != nil {
			log.Fatal(err)
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
	mux.HandleFunc("/help.html", server.CreateHelp())
	mux.Handle("/assets/", Cache(http.FileServer(http.FS(server.Assets)), 180, !*debug))
	mux.Handle("/js/execute.js", Cache(server.RunMode(*onServer), 180, !*debug))
	if *onServer {
		mux.HandleFunc("/execute/", sc.CheckSessionRestFunc(server.Execute))
	}
	mux.HandleFunc("/example/", sc.CheckSessionRestFunc(server.CreateExamples(examples)))
	mux.HandleFunc("/files/", sc.CheckSessionRestFunc(server.Files))

	serv := &http.Server{Addr: ":" + strconv.Itoa(*port), Handler: mux}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-c
		log.Print("terminated by signal ", sig.String())

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
		log.Println("Starting server with TLS at port", *port)
		err = serv.ListenAndServeTLS(*cert, *key)
	} else {
		log.Println("Starting server without TLS at port", *port)
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
