package i18n

import (
	"context"

	i18n2 "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func FormatText(ctx context.Context, msgId string) string {
	return FormatTextWithData(ctx, msgId)
}

func FormatTextWithData(ctx context.Context, msgId string) string {
	hasI18n := IsHasI18n(ctx)
	if !hasI18n {
		return ""
	}
	localizer, ok := getLocalizer(ctx)
	if !ok {
		return ""
	}
	localizeConfig := &i18n2.LocalizeConfig{
		MessageID: msgId,
	}
	localize := localizer.MustLocalize(localizeConfig)
	return localize
}

func FetchCurrentLanguageFromCtx(ctx context.Context) (*language.Tag, bool) {
	v := ctx.Value(I18nCurrentLangKey)
	if l, b := v.(language.Tag); b {
		return &l, true
	}
	return nil, false
}

func LocalizedString(ctx context.Context, defaultValue string, langMap map[language.Tag]string) string {
	langTag, tagExists := FetchCurrentLanguageFromCtx(ctx)
	if !tagExists {
		return defaultValue
	}
	str, ok := langMap[*langTag]
	if !ok {
		return defaultValue
	}
	return str
}
