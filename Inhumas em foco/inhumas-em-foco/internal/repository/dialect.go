package repository

import (
	"context"
	"strings"
)

func (r *Repository) insertID(ctx context.Context, query string, args ...any) (int64, error) {
	if r.driver == "postgres" {
		var id int64
		query = strings.TrimSpace(query)
		if !strings.Contains(strings.ToLower(query), " returning ") {
			query += " RETURNING id"
		}
		if err := r.db.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
			return 0, err
		}
		return id, nil
	}
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}
