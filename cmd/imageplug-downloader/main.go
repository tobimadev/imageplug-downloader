package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

type server struct {
	//router *http.ServeMux
	httpClient  *http.Client
	imageTokens chan bool
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
		httpClient:  &http.Client{Timeout: time.Second * 10},
		imageTokens: make(chan bool, 4),
		//router: http.NewServeMux(),
	}
	//log.Printf("srv=%+v\n", srv)

	// todo: use context

	name := time.Now().Format("060102_150405")

	if err := srv.doReport("./downloads", name, *url); err != nil {
		log.Fatalf("Error. err=%+v\n", err)
		return
	}

	// products, err := srv.readReport(*url)
	// if err != nil {
	// 	log.Printf("Error. Could not read download link. err=%+v\n", err)
	// 	return
	// }
	// log.Printf("products=%+v\n", products)

	var wg sync.WaitGroup

	// wg.Add(1)
	// httpSrv := &http.Server{Addr: ":" + *port, Handler: srv.router}
	// go func() {
	// 	defer wg.Done()
	// 	if err := httpSrv.ListenAndServe(); err != http.ErrServerClosed {
	// 		clients.Lw.Err("http-ListenAndServe", clients.Lw.Wrap(err), nil)
	// 	}
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	products, err := srv.readReport(*url)
	// 	if err != nil {
	// 		log.Printf("Error. Could not read download link. err=%+v\n", err)
	// 		return
	// 	}
	// 	log.Printf("products=%+v\n", products)
	// }()

	wg.Wait()

	// select {
	// case signal := <-stopChan:
	//     fmt.Println("received message", msg)
	// case sig := <-signals:
	//     fmt.Println("received signal", sig)
	// default:
	//     fmt.Println("no activity")
	// }

	//signal := <-stopChan
	//log.Printf("got signal=%+v\n", signal)
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

const (
	filePerm = 0666
	dirPerm  = 0755
)

func (srv *server) doReport(rootDir, name, url string) error {

	if exists, isDir := fileStatus(rootDir); !exists || !isDir {
		return fmt.Errorf("Root directory '%s' does not exist", rootDir)
	}

	reportDir := filepath.Join(rootDir, name)
	if exists, _ := fileStatus(reportDir); exists {
		return fmt.Errorf("Report directory '%s' already exist", reportDir)
	}

	products, err := srv.readReport(url)
	if err != nil {
		log.Printf("Error. Could not read download link. err=%+v\n", err)
		return err
	}
	//log.Printf("products=%+v\n", products)

	if err := os.Mkdir(reportDir, dirPerm); err != nil {
		return fmt.Errorf("Report directory '%s' could not be created. err=%+v", reportDir, err)
	}
	log.Printf("Created report dir: %s\n", reportDir)
	productTokens := make(chan bool, 3)

	var wg sync.WaitGroup

	for i := range products {

		productDir := filepath.Join(reportDir, products[i].Handle)
		if err := os.Mkdir(productDir, dirPerm); err != nil {
			log.Printf("Could not create product dir '%s'; err=%+v\n", productDir, err)
			continue
		}

		productTokens <- true
		if err := srv.downloadProduct(products[i], productDir); err != nil {
			log.Printf("Could not download product '%s'; err=%+v\n", products[i].Handle, err)
		}
		<-productTokens

		//log.Printf("Created product dir: %s\n", productDir)
		// for _, img := range p.Images {
		// 	tokens <- true
		// 	log.Printf("tokens=%d\n", len(tokens))
		// 	wg.Add(1)
		// 	go func(dir, src string) {
		// 		if err := srv.downloadImage(dir, src); err != nil {
		// 			log.Printf("Could not download image '%s'; err=%+v\n", src, err)
		// 		}
		// 		wg.Done()
		// 		<-tokens
		// 	}(productDir, img.Src)

		// 	//<-tokens
		// }
	}

	wg.Wait()

	return nil
}

// type imageManifest struct {
// 	ID       int64  `json:"id"`
// 	Src      string `json:"src"`
// 	Filename string `json:"file"`
// 	Hash     string `json:"hash"`
// }

// type prodManifest struct {
// 	ProductID int64           `json:"productId"`
// 	Handle    string          `json:"productHandle"`
// 	Created   time.Time       `json:"created"`
// 	Images    []imageManifest `json:"images"`
// }

type manifest struct {
	ProductID int64     `json:"productId"`
	Handle    string    `json:"productHandle"`
	Created   time.Time `json:"created"`
	Images    []image   `json:"images"`
}

func (srv *server) downloadProduct(p product, productDir string) error {
	var wg sync.WaitGroup

	for i := range p.Images {
		srv.imageTokens <- true
		log.Printf("tokens=%d\n", len(srv.imageTokens))
		wg.Add(1)
		go func(dir string, img *image) {
			if err := srv.downloadImage(img, dir); err != nil {
				img.Err = err
				log.Printf("Could not download image '%s'; err=%+v\n", img.Src, err)
			}
			log.Printf("image hash: %s\n", img.Hash)
			wg.Done()
			<-srv.imageTokens
		}(productDir, &p.Images[i])
	}
	wg.Wait()

	// todo: check file errors
	manifestImages := make([]image, 0, len(p.Images))
	for _, img := range p.Images {
		if img.Err != nil {
			continue
		}
		manifestImages = append(manifestImages, img)
	}
	prodManifest := manifest{ProductID: p.ID, Handle: p.Handle, Created: time.Now().UTC(), Images: manifestImages}

	json, err := json.MarshalIndent(prodManifest, "", "  ")
	if err != nil {
		return err
	}
	manifestPath := filepath.Join(productDir, fmt.Sprintf("imageplug-%d.json", p.ID))
	if err := ioutil.WriteFile(manifestPath, json, filePerm); err != nil {
		log.Printf("Could not write manifest file '%s', err=%+v\n", manifestPath, err)
	}
	//log.Printf("manifest: %s\n", string(json))
	return nil
	// todo: create manifest
}

func (srv *server) downloadImage(img *image, productDir string) error {
	log.Printf("start download: %s\n", img.Src)
	var err error
	img.Filename, err = srcToFilename(img.Src)
	if err != nil {
		return err
	}
	//log.Printf("filename=%s\n", filename)

	resp, err := srv.httpClient.Get(img.Src)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	img.Hash = toHash(buff)

	imagePath := path.Join(productDir, img.Filename)

	//log.Printf("imagePath=%s\n", imagePath)
	//f, err := os.Create(imagePath)
	//if err != nil {
	//	return err
	//}
	//defer f.Close()

	return ioutil.WriteFile(imagePath, buff, filePerm)

	// _, err = io.Copy(f, resp.Body)
	// if err != nil {
	// 	return err
	// }

	//log.Printf("Downloaded %s\n", imagePath)
	//log.Printf("done download: %s\n", img.Src)
	//return nil
}

func srcToFilename(srcURL string) (string, error) {
	src, err := url.Parse(strings.ReplaceAll(srcURL, `\/`, `/`))
	if err != nil {
		return "", err
	}
	filename := path.Base(src.Path)
	return filename, nil
}

func fileStatus(path string) (bool, bool) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, false
	}
	return true, fi.IsDir()
}

// func isDir(path string) {
// 	fi, err := file.Stat()
// 	if err != nil {
//     	// handle the error and return
// 	}
// 	if fi.IsDir() {
//     	// it's a directory
// 	} else {
//     	// it's not a directory
// 	}
// }

func toHash(data []byte) string {
	h := sha256.New()
	h.Write(data)
	hash := fmt.Sprintf("%x", h.Sum(nil))
	return hash
	//return hash[0:32]
}
