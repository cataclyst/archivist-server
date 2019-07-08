package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("Hello World")

	archivistHandler := &ArchivistHandler{}

	server := &http.Server{
		Addr:           ":8080",
		Handler:        archivistHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(server.ListenAndServe())
}

type ArchivistHandler struct{}

func (*ArchivistHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	time.Sleep(1 * time.Second)
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var arrayResult []ArchivistDocument

	for i := 0; i <= 10; i++ {
		arrayResult = append(arrayResult, ArchivistDocument{
			ID:          "abcdef" + string(i),
			Name:        "Some name " + string(i),
			Date:        time.Now(),
			Description: "Some desc",
			Labels:      []string{"label1", "label2"},
		})
	}

	result, err := json.Marshal(arrayResult)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	fmt.Fprintf(w, string(result))
}

type ArchivistDocument struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Labels      []string  `json:"labels"`
}
