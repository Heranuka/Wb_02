package pg

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"test_18/internal/config"
	"test_18/internal/domain"
	"test_18/pkg/e.go"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	logger *slog.Logger
	pool   *pgxpool.Pool
}

func NewPostgres(ctx context.Context, cfg config.Config, logger *slog.Logger) (*Postgres, error) {
	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Database, cfg.Postgres.SSLMode,
	)
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, e.Wrap("storage.pg.NewPostgres.ParseConfig", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, e.Wrap("storage.pg.NewPostgres.NewWithConfig", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, e.Wrap("storage.pg.NewPostgres.Ping", err)
	}

	return &Postgres{
		logger: logger,
		pool:   pool,
	}, nil
}

func (p *Postgres) CreateEvent(ctx context.Context, event domain.Event) (int, error) {
	var id int
	query := `INSERT INTO events (title, description, start_time, 
	end_time, location, created_at, updated_at) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err := p.pool.QueryRow(ctx, query, event.Title, event.Description, event.StartTime, event.EndTime, event.Location, event.CreatedAt, event.UpdatedAt).Scan(&id)
	if err != nil {
		return 0, e.Wrap("storage.pg.CreateEvent", err)
	}

	return id, nil
}

func (p *Postgres) UpdateEvent(ctx context.Context, id int, req domain.UpdateEventRequest) error {
	var event domain.Event

	query := `UPDATE events SET 
			 title = COALESCE($2, title),
            description = COALESCE($3, description),
            start_time = COALESCE($4, start_time),
            end_time = COALESCE($5, end_time),
            location = COALESCE($6, location),
            updated_at = NOW()
			WHERE id = $1 RETURNING id, title, description, start_time, end_time, location, created_at, updated_at `

	err := p.pool.QueryRow(ctx, query, id, req.Title, req.Description, req.StartTime, req.EndTime, req.Location).Scan(&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime, &event.Location, &event.CreatedAt, &event.UpdatedAt)
	if err != nil {
		return e.Wrap("storage.pg.UpdateEvent", err)
	}

	return nil
}

func (p *Postgres) DeleteEvent(ctx context.Context, id int) error {
	query := `DELETE FROM events WHERE id = $1`
	_, err := p.pool.Exec(ctx, query, id)
	if err != nil {
		return e.Wrap("storage.pg.DeleteEvent", err)
	}
	return nil
}
func (p *Postgres) GetEventsForTime(ctx context.Context, startDate time.Time, endDate time.Time) ([]domain.Event, error) {
	var events []domain.Event
	query := `SELECT id, title, description, start_time, 
	end_time, location, created_at, updated_at FROM events WHERE start_time >= $1 AND start_time < $2`

	rows, err := p.pool.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, e.Wrap("storage.pg.GetEventsForDay", err)
	}
	defer rows.Close()

	for rows.Next() {
		var event domain.Event
		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.StartTime,
			&event.EndTime,
			&event.Location,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			log.Printf("GetEventsForTime: Scan error: %v", err)
			return nil, e.Wrap("storage.pg.GetEventsForTime", err)
		}

		log.Printf("GetEventsForTime: Event - ID: %d Title: %s, StartTime: %v", event.ID, event.Title, event.StartTime) // Добавляем логирование
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, e.Wrap("storage.pg.GetEventsForDay", err)
	}

	return events, nil
}

func (p *Postgres) CloseDB() {
	p.pool.Close()
	stat := p.pool.Stat()
	if stat.AcquiredConns() > 0 {
		p.logger.Warn("postgres connections not fully closed after Close()", slog.Any("acquired connections", stat.AcquiredConns()))

	}
}
