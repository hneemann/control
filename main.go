package main

import (
	"context"
	"errors"
	"flag"
	"github.com/hneemann/control/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
)

func main() {
	//folder := flag.String("folder", "/home/hneemann/Dokumente/Fahrrad/Gefahren/tracks", "track folder")
	cert := flag.String("cert", "localhost.pem", "certificate")
	key := flag.String("key", "localhost.key", "certificate")
	port := flag.Int("port", 8080, "port")
	flag.Parse()

	examples := server.ReadExamples()

	http.Handle("/assets/", http.FileServer(http.FS(server.Assets)))
	http.HandleFunc("/", server.CreateMain(examples))
	http.HandleFunc("/execute/", server.Execute)
	http.HandleFunc("/example/", server.CreateExamples(examples))

	serv := &http.Server{Addr: ":" + strconv.Itoa(*port)}

	shutdownAck := make(chan struct{})
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Print("terminated")
		err := serv.Shutdown(context.Background())
		if err != nil {
			log.Println(err)
		}
		//server.ShutdownDB()
		close(shutdownAck)
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
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("waiting for table shutdown")
		<-shutdownAck
	}

}
