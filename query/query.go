package query

import (
	"errors"

	"github.com/ikwerre-dev/emojidb/core"
)

type Query struct {
	Db        *core.Database
	TableName string
	Filters   []FilterFunc
	Columns   []string
}

type FilterFunc func(core.Row) bool

func NewQuery(db *core.Database, tableName string) *Query {
	return &Query{
		Db:        db,
		TableName: tableName,
	}
}

func (q *Query) Filter(f FilterFunc) *Query {
	q.Filters = append(q.Filters, f)
	return q
}

func (q *Query) Select(columns ...string) *Query {
	q.Columns = columns
	return q
}

func (q *Query) Execute() ([]core.Row, error) {
	q.Db.Mu.RLock()
	table, ok := q.Db.Tables[q.TableName]
	q.Db.Mu.RUnlock()

	if !ok {
		return nil, errors.New("table not found: " + q.TableName)
	}

	var results []core.Row

	table.Mu.RLock()
	for _, row := range table.HotHeap.Rows {
		if q.Matches(row) {
			results = append(results, q.Project(row))
		}
	}

	for _, clump := range table.SealedClumps {
		for _, row := range clump.Rows {
			if q.Matches(row) {
				results = append(results, q.Project(row))
			}
		}
	}
	table.Mu.RUnlock()

	return results, nil
}

func (q *Query) Matches(row core.Row) bool {
	for _, filter := range q.Filters {
		if !filter(row) {
			return false
		}
	}
	return true
}

func (q *Query) Project(row core.Row) core.Row {
	if len(q.Columns) == 0 {
		return row
	}

	projected := make(core.Row)
	for _, col := range q.Columns {
		if val, ok := row[col]; ok {
			projected[col] = val
		}
	}
	return projected
}
