package i18n

import "golang.org/x/text/language"

func MatchCurrentLanguageTag(accept string, supportTags []language.Tag) language.Tag {
	langTags, _, err := language.ParseAcceptLanguage(accept)
	if err != nil {
		langTags = []language.Tag{language.Chinese}
	}
	var matcher = language.NewMatcher(supportTags)
	_, i, _ := matcher.Match(langTags...)
	tag := supportTags[i]
	if tag == language.Und {
		tag = language.Chinese
	}
	return tag
}
