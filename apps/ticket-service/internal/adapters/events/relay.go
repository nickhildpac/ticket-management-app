package events

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// StreamKey is the Redis stream the AI worker consumes.
const StreamKey = "ticket-events"

// Relay polls the event_outbox table and publishes unpublished rows to a Redis
// stream, then marks them published. At-least-once: a crash between XADD and the
// UPDATE re-delivers the row, so consumers must be idempotent (dedupe on event id).
type Relay struct {
	db       *sql.DB
	rdb      *redis.Client
	interval time.Duration
	batch    int
}

func NewRelay(db *sql.DB, rdb *redis.Client) *Relay {
	return &Relay{db: db, rdb: rdb, interval: time.Second, batch: 100}
}

// Run blocks, draining the outbox until ctx is cancelled.
func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if n, err := r.drainOnce(ctx); err != nil {
				log.Printf("outbox relay: %v", err)
			} else if n > 0 {
				log.Printf("outbox relay: published %d event(s)", n)
			}
		}
	}
}

func (r *Relay) drainOnce(ctx context.Context) (int, error) {
	const sel = `SELECT id, event_type, aggregate_id, payload
	             FROM event_outbox
	             WHERE published_at IS NULL
	             ORDER BY created_at
	             LIMIT $1`
	rows, err := r.db.QueryContext(ctx, sel, r.batch)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type outboxRow struct {
		id, eventType, aggregateID string
		payload                    []byte
	}
	var pending []outboxRow
	for rows.Next() {
		var row outboxRow
		if err := rows.Scan(&row.id, &row.eventType, &row.aggregateID, &row.payload); err != nil {
			return 0, err
		}
		pending = append(pending, row)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	published := 0
	for _, row := range pending {
		values := map[string]any{
			"event_id":     row.id,
			"event_type":   row.eventType,
			"aggregate_id": row.aggregateID,
			"payload":      string(row.payload),
		}
		if err := r.rdb.XAdd(ctx, &redis.XAddArgs{Stream: StreamKey, Values: values}).Err(); err != nil {
			return published, err
		}
		if _, err := r.db.ExecContext(ctx,
			`UPDATE event_outbox SET published_at = now() WHERE id = $1`, row.id); err != nil {
			return published, err
		}
		published++
	}
	return published, nil
}
