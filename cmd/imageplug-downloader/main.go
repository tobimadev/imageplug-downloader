package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type server struct {
	//router *http.ServeMux
	httpClient *http.Client
}

// func getEnvDef(name string, def string) string {
// 	v := os.Getenv(name)
// 	if v == "" {
// 		return def
// 	}
// 	return v
// }

func main() {
	rand.Seed(time.Now().UnixNano())
	stopChan := make(chan os.Signal, 3)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	url := flag.String("url", "", "download link copied from ImagePlug app")
	//gcpProject := flag.String("gcpProj", getEnvDef("gcpProj", "image-init-dev"), "gcp project")
	//gcpLocation := flag.String("gcpLocation", getEnvDef("gcpLocation", "us-east4"), "gcp location")
	//prof := flag.String("prof", getEnvDef("prof", "dev"), "profile, dev/stage/prod")
	//gcpBucket := flag.String("gcpBucket", getEnvDef("gcpBucket", "tob-image-init-dev"), "gcp bucket")

	flag.Parse()

	log.Printf("url=%s\n", *url)

	srv := server{
		httpClient: &http.Client{Timeout: time.Second * 10},
		//router: http.NewServeMux(),
	}
	//log.Printf("srv=%+v\n", srv)

	// todo: use context

	var wg sync.WaitGroup

	// wg.Add(1)
	// httpSrv := &http.Server{Addr: ":" + *port, Handler: srv.router}
	// go func() {
	// 	defer wg.Done()
	// 	if err := httpSrv.ListenAndServe(); err != http.ErrServerClosed {
	// 		clients.Lw.Err("http-ListenAndServe", clients.Lw.Wrap(err), nil)
	// 	}
	// }()

	wg.Add(1)
	go func() {
		defer wg.Done()
		products, err := srv.readReport(*url)
		if err != nil {
			log.Printf("Error. Could not read download link. err=%+v\n", err)
			return
		}
		log.Printf("products=%+v\n", products)
	}()

	signal := <-stopChan
	log.Printf("got signal=%+v\n", signal)
	// clients.Lw.Event("got-kill-signal", map[string]interface{}{"signal": signal})
	// cancelCtx()

	// ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancelFunc()
	// if err := httpSrv.Shutdown(ctx); err != nil {
	// 	clients.Lw.Err("http-Shutdown", clients.Lw.Wrap(err), nil)
	// }

	//wg.Wait()
	//clients.Lw.Event("server-done", nil)
	//clients.Lw.Flush()
}
