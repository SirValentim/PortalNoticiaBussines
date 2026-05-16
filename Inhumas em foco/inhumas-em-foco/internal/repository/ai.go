package repository

import (
	"context"

	"inhumas-em-foco/internal/model"
)

func (r *Repository) AIUsageLogCreate(ctx context.Context, log *model.AIUsageLog) error {
	id, err := r.insertID(ctx, `
		INSERT INTO ai_usage_logs (post_id, user_id, action, provider, input_title, output, source_name, source_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		log.PostID, log.UserID, log.Action, log.Provider, log.InputTitle, log.Output, log.SourceName, log.SourceURL)
	if err != nil {
		return err
	}
	log.ID = id
	return nil
}

func (r *Repository) AIUsageLogListForPost(ctx context.Context, postID int64, limit int) ([]model.AIUsageLog, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT l.id, l.post_id, l.user_id, l.action, l.provider, COALESCE(l.input_title, ''), COALESCE(l.output, ''), COALESCE(l.source_name, ''), COALESCE(l.source_url, ''), l.created_at, COALESCE(u.name, ''), COALESCE(u.email, '')
		FROM ai_usage_logs l
		LEFT JOIN users u ON u.id = l.user_id
		WHERE l.post_id = $1
		ORDER BY l.created_at DESC
		LIMIT $2`, postID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []model.AIUsageLog
	for rows.Next() {
		var log model.AIUsageLog
		if err := rows.Scan(&log.ID, &log.PostID, &log.UserID, &log.Action, &log.Provider, &log.InputTitle, &log.Output, &log.SourceName, &log.SourceURL, &log.CreatedAt, &log.UserName, &log.UserEmail); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}
