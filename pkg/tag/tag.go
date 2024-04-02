package tag

import (
	"context"
	"errors"

	"github.com/ServiceWeaver/weaver"
	"github.com/pedromol/camelo/pkg/database"
	"github.com/pedromol/camelo/pkg/homeassistant"
	"github.com/pedromol/camelo/pkg/model"
)

type Service interface {
	Get(ctx context.Context, value string) error
	Post(ctx context.Context, value model.Tag) error
}

type Tag struct {
	weaver.Implements[Service]
	Hass weaver.Ref[homeassistant.HomeAssistant]
	Db   weaver.Ref[database.Database]
}

func (t Tag) Get(ctx context.Context, value string) error {
	if value == "" {
		return errors.New("missing value")
	}

	v, err := t.Db.Get().Get(ctx, value)
	if err != nil {
		return errors.New("not found")
	}

	err = t.Hass.Get().Toggle(ctx, v.Device)
	if err != nil {
		return errors.New("failed to toggle")
	}
	return nil
}

func (t Tag) Post(ctx context.Context, value model.Tag) error {
	err := t.Db.Get().Upsert(ctx, value)
	if err != nil {
		return errors.New("failed to upsert")
	}
	return nil
}
