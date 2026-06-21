package app

import "github.com/stuffstash/stuff-stash/internal/app/appsupport"

const paginationCursorVersion = appsupport.PaginationCursorVersion

func pageLimit(defaultLimit int, maxLimit int, requested int) int {
	return appsupport.PageLimit(defaultLimit, maxLimit, requested)
}

func encodePageCursor(collection string, scope string, lastID string) *string {
	return appsupport.EncodePageCursor(collection, scope, lastID)
}

func decodePageCursor(collection string, scope string, cursor string) (string, error) {
	return appsupport.DecodePageCursor(collection, scope, cursor)
}
