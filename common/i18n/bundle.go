package i18n

import (
	i18n2 "github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
)

func NewBundle(tag language.Tag, configs ...string) *i18n2.Bundle {
	bundle := i18n2.NewBundle(tag)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	for _, file := range configs {
		_, err := bundle.LoadMessageFile(file)
		if err != nil {
			panic(err)
		}
	}
	return bundle
}
