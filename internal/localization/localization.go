package localization

import (
	"context"
	"fmt"
	"net/http"
	"os"

	envvars "github.com/mkaykisiz/sender/configs/env-vars"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// global bundle
var bundle *i18n.Bundle

// default language
var defaultLanguage = language.Turkish

// Localizer defines behaviors of localizer
type Localizer interface {
	Localize(l *i18n.Localizer)
}

// InitializeBundle initializes bundle
func InitializeBundle(l envvars.Localization) error {
	bundle = i18n.NewBundle(defaultLanguage)

	languageFiles, err := os.ReadDir(l.LanguageFilesDirectory)
	if err != nil {
		return fmt.Errorf("reading language files directory failed, %s", err.Error())
	}

	for _, languageFile := range languageFiles {
		if languageFile.IsDir() == true {
			continue
		}

		_, err := bundle.LoadMessageFile(fmt.Sprintf("%s/%s", l.LanguageFilesDirectory, languageFile.Name()))
		if err != nil {
			return fmt.Errorf("loading language (message) file failed, %s", err.Error())
		}
	}

	return nil
}

var localizerKey = struct{ Key string }{"localizer"}

// AddLocalizerToContext creates and adds new localizer to context
func AddLocalizerToContext(ctx context.Context, r *http.Request) context.Context {
	acceptLanguage := r.Header.Get("Accept-Language")

	l := i18n.NewLocalizer(bundle, acceptLanguage)

	return context.WithValue(ctx, localizerKey, *l)
}

// GetLocalizerFromContext gets and returns localizer from context
func GetLocalizerFromContext(ctx context.Context) *i18n.Localizer {
	l, ok := ctx.Value(localizerKey).(i18n.Localizer)
	if !ok {
		return nil
	}

	return &l
}

// Localize returns localized message or default message
func Localize(l *i18n.Localizer, messageID string, defaultMessage string) string {
	localizedMessage, err := l.Localize(&i18n.LocalizeConfig{
		MessageID: messageID,
	})
	if err != nil {
		return defaultMessage
	}

	return localizedMessage
}

// LocalizeWithConfig returns localized message or default message with config
func LocalizeWithConfig(l *i18n.Localizer, messageID string, templateData interface{}, pluralCount interface{}, defaultMessage string) string {
	localizedMessage, err := l.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
		PluralCount:  pluralCount,
	})
	if err != nil {
		return defaultMessage
	}

	return localizedMessage
}
