// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/tickraft/tickraft/pkg/api/httputil"
	"github.com/tickraft/tickraft/pkg/i18n"
)

// localeHeader is the HTTP header that carries the request locale as a BCP 47
// language tag. It is sent by the frontend on every request so the backend can
// render localized alert messages, format logs, and drive any locale-aware
// response formatting.
const localeHeader = "X-Tickraft-Locale"

// NewLocaleMiddleware returns a Hertz middleware that reads the
// X-Tickraft-Locale header from each request, parses it as a BCP 47 locale
// tag, and stores the resulting i18n.Locale in the request context for
// downstream handlers.
//
// When the header is absent or invalid, i18n.Parse falls back to the default
// locale (zh-Hans), so downstream handlers always receive a usable Locale value.
//
// Downstream handlers retrieve the locale via LocaleFromContext or
// httputil.GetLocale.
func NewLocaleMiddleware() app.HandlerFunc {
	return func(ctx context.Context, arc *app.RequestContext) {
		header := string(arc.GetHeader(localeHeader))
		loc := i18n.Parse(header)
		ctx = httputil.SetLocale(ctx, loc)
		arc.Next(ctx)
	}
}

// LocaleFromContext extracts the locale from the request context. It returns
// the default locale (zh-Hans) when no locale has been set, ensuring callers
// always receive a usable Locale value.
func LocaleFromContext(ctx context.Context) i18n.Locale {
	return httputil.GetLocale(ctx)
}
