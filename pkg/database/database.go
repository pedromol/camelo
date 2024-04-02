package database

import (
	"context"

	"github.com/pedromol/camelo/pkg/model"
)

type Database interface {
	Init(context.Context) error
	Get(context.Context, string) (model.Tag, error)
	Upsert(context.Context, model.Tag) error
}
