package alerts

import (
	"context"
	"database/sql"
	"time"
)

func CountBySeveritySince(ctx context.Context, db *sql.DB, since time.Time) (map[string]int, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT severity, COUNT(1) FROM alert_events WHERE created_at >= $1 GROUP BY severity`,
		since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]int)
	for rows.Next() {
		var severity string
		var count int
		if err := rows.Scan(&severity, &count); err != nil {
			return nil, err
		}
		out[severity] = count
	}
	return out, rows.Err()
}
