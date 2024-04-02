package homeassistant

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/ServiceWeaver/weaver"
)

type HomeAssistant interface {
	Init(context.Context) error
	Toggle(context.Context, string) error
}

type HASS struct {
	weaver.Implements[HomeAssistant]
	auth string
	host string
}

func (h *HASS) Init(ctx context.Context) error {
	auth := os.Getenv("HOMEASSISTANT_AUTH")
	if auth == "" {
		log.Println("HOMEASSISTANT_AUTH not set")
		return errors.New("HOMEASSISTANT_AUTH not set")
	}
	h.auth = auth

	host := os.Getenv("HOMEASSISTANT_HOST")
	if host == "" {
		log.Println("HOMEASSISTANT_HOST not set")
		return errors.New("HOMEASSISTANT_HOST not set")
	}
	h.host = host
	return nil
}

func (h HASS) Toggle(ctx context.Context, entityIDs string) error {
	wg := sync.WaitGroup{}

	for _, entityID := range strings.Split(entityIDs, ",") {
		wg.Add(1)
		go func(entityID string, wg *sync.WaitGroup) {
			defer wg.Done()
			log.Printf("toggling %s\n", entityID)
			req, err := http.NewRequestWithContext(
				ctx,
				http.MethodPost,
				h.host+"/api/services/light/toggle",
				strings.NewReader(fmt.Sprintf(`{"entity_id": "%s"}`, entityID)),
			)
			if err != nil {
				log.Printf("error creating request: %v\n", err)
				return
			}
			req.Header.Set("Authorization", h.auth)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("error posting: %v\n", err)
				return
			}
			resp.Body.Close()
		}(entityID, &wg)
	}
	wg.Wait()
	return nil
}
