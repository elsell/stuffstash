package gormstore

import (
	"context"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const searchCandidateBatchSize = 128

// postgresSearchCandidates is deliberately a superset: domain matching still
// decides equality, match labels, and Go's custom-field/Unicode representation.
func (s Store) postgresSearchCandidates(ctx context.Context, query *gorm.DB, tenantID tenant.ID, inventories []string, page ports.AssetSearchPageRequest) *gorm.DB {
	text := strings.TrimSpace(strings.ToLower(page.Query.String()))
	if s.db.Dialector.Name() != "postgres" || text == "" {
		return query
	}
	for _, character := range text {
		if character == 0 || character > 127 {
			return query
		}
	}
	// Every domain match contains every term, so one indexed term is a safe
	// superset. Prefer the longest without multiplying metadata subqueries.
	text = ""
	for _, term := range page.Query.Terms(page.Mode) {
		if len(term) > len(text) {
			text = term
		}
	}
	pattern := "%" + strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(text) + "%"
	scoped := func(model any) *gorm.DB {
		return s.db.WithContext(ctx).Model(model).Where(map[string]any{"tenant_id": tenantID.String(), "inventory_id": inventories})
	}
	attachments := scoped(&attachmentModel{}).Select("asset_id").Where(searchTextCandidates(pattern, "file_name", "content_type"))
	types := s.db.WithContext(ctx).Model(&customAssetTypeModel{}).Select("id").Where(map[string]any{"tenant_id": tenantID.String()}).Where(searchTextCandidates(pattern, "type_key", "display_name", "description"))
	tags := scoped(&assetTagModel{}).Select("id").Where(map[string]any{"lifecycle_state": "active"}).Where(searchTextCandidates(pattern, "key", "display_name"))
	assignments := scoped(&assetTagAssignmentModel{}).Select("asset_id").Where(searchInSubquery("tag_id", tags))
	return query.Where(clause.Or(
		searchTextCandidates(pattern, "title", "description"),
		clause.Neq{Column: clause.Column{Name: "custom_fields"}, Value: "{}"},
		searchInSubquery("id", attachments),
		searchInSubquery("custom_asset_type_id", types),
		searchInSubquery("id", assignments),
	))
}

func searchTextCandidates(pattern string, columns ...string) clause.Expression {
	expressions := make([]clause.Expression, 0, len(columns))
	for _, name := range columns {
		column := clause.Column{Name: name}
		// Multibyte values stay candidates even if database case folding differs
		// from Go. All column names are repository-owned; values are parameters.
		expressions = append(expressions, clause.Expr{SQL: `translate(?, 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz') COLLATE "C" LIKE ? OR octet_length(?) <> char_length(?)`, Vars: []any{column, pattern, column, column}})
	}
	return clause.Or(expressions...)
}

func searchInSubquery(column string, query *gorm.DB) clause.Expression {
	return clause.Expr{SQL: "? IN (?)", Vars: []any{clause.Column{Name: column}, query}}
}

// CursorKey is a bytewise concatenated key, not a tuple: variable-length IDs
// must preserve the same order as the existing Go comparison and opaque cursors.
func searchCursorExpression(dialect string) clause.Expr {
	collation := "BINARY"
	if dialect == "postgres" {
		collation = `"C"`
	}
	return clause.Expr{SQL: "(? || ':' || ?) COLLATE " + collation, Vars: []any{clause.Column{Name: "inventory_id"}, clause.Column{Name: "id"}}}
}

func searchAfterCursor(query *gorm.DB, cursor string) *gorm.DB {
	expression := searchCursorExpression(query.Dialector.Name())
	query = query.Clauses(clause.OrderBy{Expression: expression})
	if cursor == "" {
		return query
	}
	return query.Where(clause.Gt{Column: expression, Value: cursor})
}
