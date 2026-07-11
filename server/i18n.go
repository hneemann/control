package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"net/http"
	"path/filepath"
)

type Translator struct {
	*message.Printer
}

func (t Translator) MsgID(key string, args ...any) string {
	return t.Sprintf(key, args...)
}

type TranslatorFactory func(request *http.Request) Translator

func LoadJSONTranslations(f embed.FS, path string, defLang language.Tag) (TranslatorFactory, error) {
	entries, err := f.ReadDir(path)
	if err != nil {
		return nil, err
	}
	if len(entries) < 1 {
		return nil, fmt.Errorf("expected at least one files in %s", path)
	}

	var tags []language.Tag
	for _, entry := range entries {
		data, err := TemplateFS.ReadFile(filepath.Join(path, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("could not read file %s: %w", entry.Name(), err)
		}

		var translations map[string]string
		if err = json.Unmarshal(data, &translations); err != nil {
			return nil, fmt.Errorf("could not parse JSON file %s: %w", entry.Name(), err)
		}

		id, ok := translations["langID"]
		if !ok {
			return nil, fmt.Errorf("could not find key langID in file %s", entry.Name())
		}
		tag, err := language.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("could not parse language id %s: %w", id, err)
		}
		tags = append(tags, tag)

		for key, value := range translations {
			err = message.SetString(tag, key, value)
			if err != nil {
				return nil, fmt.Errorf("could not set translation field %s: %w", key, err)
			}
		}
	}

	// set english as the default language
	found := -1
	for i, tag := range tags {
		if tag == defLang {
			found = i
			break
		}
	}
	if found == -1 {
		return nil, fmt.Errorf("could not find a translation tag for the default language")
	}
	if found != 0 {
		tags[0], tags[found] = tags[found], tags[0]
	}

	return createTranslatorFactory(tags)
}

func createTranslatorFactory(tags []language.Tag) (TranslatorFactory, error) {
	matcher := language.NewMatcher(tags)
	return func(request *http.Request) Translator {
		accept := request.Header.Get("Accept-Language")
		tag, _ := language.MatchStrings(matcher, accept)
		p := message.NewPrinter(tag)
		return Translator{p}
	}, nil
}
