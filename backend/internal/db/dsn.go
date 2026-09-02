package db

import (
	"fmt"
	"net/url"
	"strings"
)

// dsn normalizes a connection string into a "file:" URI carrying the given
// query parameters and per-connection PRAGMAs.
//
// PRAGMAs must travel in the DSN rather than being executed after opening: a
// PRAGMA is per-connection state, so a statement run against the pool reaches
// only whichever connection database/sql happened to hand out. The driver
// (modernc.org/sqlite) applies "_pragma" parameters to every new connection.
//
// Parameters already present in connStr are preserved and take precedence
// over params, so an operator can override a default by spelling it out.
//
// bareEscaper percent-escapes the characters SQLite's URI parser would
// otherwise consume. '%' must be listed so an operator's literal '%' does not
// become an escape sequence; strings.Replacer scans once and never re-examines
// what it inserted, so the order here is safe.
var bareEscaper = strings.NewReplacer("%", "%25", "#", "%23", "?", "%3F")

func dsn(connStr string, params map[string]string, pragmas []string) (string, error) {
	raw, isURI := strings.CutPrefix(connStr, "file:")

	var path, rawQuery string
	if isURI {
		path, rawQuery, _ = strings.Cut(raw, "?")
	} else {
		// A bare path is a filename, not a URI. SQLite would otherwise
		// percent-decode it and truncate it at '#', silently opening a
		// different file than the operator named. The whole string is the
		// filename, so it is never split on '?' either.
		path = bareEscaper.Replace(raw)
	}

	q, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "", fmt.Errorf("parse dsn query %q: %w", rawQuery, err)
	}

	for k, v := range params {
		if !q.Has(k) {
			q.Set(k, v)
		}
	}
	for _, p := range pragmas {
		q.Add("_pragma", p)
	}

	if len(q) == 0 {
		return "file:" + path, nil
	}
	return "file:" + path + "?" + q.Encode(), nil
}
