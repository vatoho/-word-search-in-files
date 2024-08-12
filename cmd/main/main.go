package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"word-search-in-files/pkg/delivery"

	searcher2 "word-search-in-files/pkg/searcher"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("error in logger initialization: %s", err)
	}
	sugaredLogger := logger.Sugar()
	defer func() {
		err = sugaredLogger.Sync()
		if err != nil {
			fmt.Printf("error in logger synchronization: %s\n", err)
		}
	}()

	flagFSRoot := flag.String("f", "./", "target file system root")
	flagDir := flag.String("d", "", "target directory in file system root")
	flag.Parse()

	fSys := os.DirFS(*flagFSRoot)

	searcher := searcher2.NewSearcher(fSys, *flagDir)
	searcherHandler := delivery.NewSearcherHandler(searcher, sugaredLogger)

	err = searcher.ConstructFileDictionary()
	if err != nil {
		sugaredLogger.Fatalf("error in constructing files dictionary: %s", err)
		return
	}

	router := mux.NewRouter()
	router.HandleFunc("/files/search", searcherHandler.Search).Methods(http.MethodGet)

	port := ":8000"
	sugaredLogger.Infof("starting server on %s", port)
	sugaredLogger.Infof("Server is working on api /files/search?keyword={keyword}")

	err = http.ListenAndServe(port, router)
	if err != nil {
		sugaredLogger.Fatalf("error in starting server: %s", err)
	}

}
