package sqlutils

import (
	"fmt"
	"reflect"
	"strings"
)

type PaginatedResult[T any] struct {
	Data       []*T  `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

type QueryBuilder struct {
	table, fields, where, order string
	args                        []any
	limit, offset               int
}

func NewQueryBuilder(table string) *QueryBuilder {
	return &QueryBuilder{
		table: table,
		where: "deleted_at IS NULL",
	}
}

func (q *QueryBuilder) WithFields(fields string) *QueryBuilder {
	q.fields = fields
	return q
}

func (q *QueryBuilder) WithOrder(order string) *QueryBuilder {
	q.order = order
	return q
}

func (q *QueryBuilder) WithLimit(limit int) *QueryBuilder {
	q.limit = limit
	return q
}

func (q *QueryBuilder) WithOffset(offset int) *QueryBuilder {
	q.offset = offset
	return q
}

func (q *QueryBuilder) Add(cond string, val any) {
	if val == nil {
		return
	}
	if v := reflect.ValueOf(val); v.Kind() == reflect.Ptr && v.IsNil() {
		return
	}
	q.where += fmt.Sprintf(" AND "+cond, len(q.args)+1)
	q.args = append(q.args, val)
}

func (q *QueryBuilder) AddSearch(fields []string, term string) int {
	p := "%" + term + "%"
	pos := len(q.args) + 1
	parts := make([]string, len(fields))
	for i, f := range fields {
		parts[i] = fmt.Sprintf("%s ILIKE $%d", f, pos)
	}
	q.where += " AND (" + strings.Join(parts, " OR ") + ")"
	q.args = append(q.args, p)
	return pos
}

func (q *QueryBuilder) Where() string {
	return q.where
}

func (q *QueryBuilder) Args() []any {
	return q.args
}

func (q *QueryBuilder) Build() string {
	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", q.fields, q.table, q.where)
	if q.order != "" {
		sql += " ORDER BY " + q.order
	}
	if q.limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", q.limit)
	}
	if q.offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", q.offset)
	}
	return sql
}

func WrapLike(s *string) any {
	if s == nil {
		return nil
	}
	return "%" + *s + "%"
}
