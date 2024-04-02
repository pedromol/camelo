package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ServiceWeaver/weaver"
	"github.com/pedromol/camelo/pkg/model"
	"github.com/pedromol/camelo/pkg/tag"
)

type app struct {
	weaver.Implements[weaver.Main]
	service weaver.Ref[tag.Service]
	tags    weaver.Listener `weaver:"tags"`
}

func main() {
	if err := weaver.Run(context.Background(), serve); err != nil {
		log.Fatal(err)
	}
}

func serve(ctx context.Context, app *app) error {
	log.Printf("tags listener available on %v\n", app.tags)
	http.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		c, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if r.Method == http.MethodPost {
			m := model.Tag{}
			err := json.NewDecoder(r.Body).Decode(&m)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			err = app.service.Get().Post(c, m)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			return
		}
		if r.Method == http.MethodGet {
			value := r.URL.Query().Get("value")
			err := app.service.Get().Get(c, value)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	return http.Serve(app.tags, nil)
}
