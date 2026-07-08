package server

import (
	"encoding/json"
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"html/template"
	"net/http"
)

var matcher = language.NewMatcher([]language.Tag{
	language.English, // Fallback
	language.German,
})

func I18nFuncs(request *http.Request) template.FuncMap {
	accept := request.Header.Get("Accept-Language")
	tag, _ := language.MatchStrings(matcher, accept)
	p := message.NewPrinter(tag)
	return template.FuncMap{
		"tr": func(key string, args ...interface{}) string {
			return p.Sprintf(key, args...)
		},
	}
}

func LoadJSONTranslations() error {
	// Wir definieren, welche Sprachen wir erwarten und wie die Dateien heißen
	files := map[string]language.Tag{
		"templates/i18n/de.json": language.German,
		"templates/i18n/en.json": language.English,
	}

	for path, tag := range files {
		data, err := templateFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("could not read file %s: %w", path, err)
		}

		var translations map[string]string
		if err := json.Unmarshal(data, &translations); err != nil {
			return fmt.Errorf("could not parse JSON file %s: %w", path, err)
		}

		for key, value := range translations {
			message.SetString(tag, key, value)
		}
	}
	return nil
}
