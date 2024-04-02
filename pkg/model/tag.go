package model

import "github.com/ServiceWeaver/weaver"

type Tag struct {
	weaver.AutoMarshal
	Id     string `json:"id"`
	Device string `json:"device"`
}
