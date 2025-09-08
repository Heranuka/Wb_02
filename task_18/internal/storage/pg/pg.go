package pg

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"test_18/internal/config"
	"test_18/internal/domain"
	"test_18/pkg/e"

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

func (p *Postgres) CreateUser(ctx context.Context) (int, error) {
	var id int
	err := p.pool.QueryRow(ctx, `INSERT INTO users (created_at) VALUES (NOW()) RETURNING id`).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (p *Postgres) CreateEvent(ctx context.Context, event domain.Event) (int, error) {
	var id int
	query := `INSERT INTO events (user_id, title, description, event_date) 
              VALUES ($1, $2, $3, $4) RETURNING id`
	err := p.pool.QueryRow(ctx, query,
		event.UserID,
		event.Title,
		event.Description,
		time.Time(event.EventDate).In(time.UTC),
	).Scan(&id)
	if err != nil {
		return 0, e.Wrap("storage.pg.CreateEvent", err)
	}
	return id, nil
}

type EventWithUserTimestamps struct {
	domain.Event
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *Postgres) UpdateEvent(ctx context.Context, id int, req domain.UpdateEventRequest) error {
	query := `UPDATE events e
SET 
    title = COALESCE($2, e.title),
    description = COALESCE($3, e.description),
    event_date = COALESCE($4, e.event_date),
    updated_at = NOW()
FROM users u
WHERE e.id = $1 AND e.user_id = u.id
RETURNING 
    e.id, e.title, e.description, e.event_date, e.created_at, e.updated_at,
    u.created_at AS user_created_at,
    u.updated_at AS user_updated_at`

	var event EventWithUserTimestamps

	err := p.pool.QueryRow(ctx, query,
		id,
		req.Title,
		req.Description,
		startTimeToTimeUTC(req.EventDate),
	).Scan(
		&event.ID,
		&event.Title,
		&event.Description,
		&event.EventDate,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		return e.Wrap("storage.pg.UpdateEvent", err)
	}
	return nil
}

func startTimeToTimeUTC(date *domain.Date) *time.Time {
	if date == nil {
		return nil
	}
	t := time.Time(*date).In(time.UTC)
	return &t
}

func (p *Postgres) DeleteEvent(ctx context.Context, id int) error {
	query := `DELETE FROM events WHERE id = $1`
	_, err := p.pool.Exec(ctx, query, id)
	if err != nil {
		return e.Wrap("storage.pg.DeleteEvent", err)
	}
	return nil
}

func (p *Postgres) GetEventsForDay(ctx context.Context, userId int, day time.Time) ([]domain.Event, error) {
	var events []domain.Event
	query := `SELECT id, title, description, event_date, created_at, updated_at 
              FROM events WHERE user_id = $1 AND event_date = $2`

	start := day.In(time.UTC)
	end := start.Add(24 * time.Hour)

	rows, err := p.pool.Query(ctx, query, userId, start, end)
	if err != nil {
		return nil, e.Wrap("storage.pg.GetEventsByCertainDay", err)
	}
	defer rows.Close()

	for rows.Next() {
		var event domain.Event
		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.EventDate,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			log.Printf("GetEventsByCertainDay: Scan error: %v", err)
			return nil, e.Wrap("storage.pg.GetEventsByCertainDay", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, e.Wrap("storage.pg.GetEventsByCertainDay", err)
	}
	return events, nil
}

func (p *Postgres) GetEventsForWeek(ctx context.Context, userId int, date time.Time) ([]domain.Event, error) {
	var events []domain.Event
	start := date.AddDate(0, 0, -7)

	end := date.AddDate(0, 0, 1)
	query := `SELECT id, title, description, event_date, created_at, updated_at 
              FROM events WHERE user_id = $1 AND event_date >= $2 AND event_date < $3`

	rows, err := p.pool.Query(ctx, query, userId, start, end)
	if err != nil {
		return nil, e.Wrap("storage.pg.GetEventsByCertainDay", err)
	}
	defer rows.Close()

	for rows.Next() {
		var event domain.Event
		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.EventDate,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			log.Printf("GetEventsByCertainDay: Scan error: %v", err)
			return nil, e.Wrap("storage.pg.GetEventsByCertainDay", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, e.Wrap("storage.pg.GetEventsByCertainDay", err)
	}
	return events, nil
}

func (p *Postgres) GetEventsForMonth(ctx context.Context, userId int, date time.Time) ([]domain.Event, error) {
	var events []domain.Event

	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	end := date.AddDate(0, 1, 0)

	query := `SELECT id, title, description, event_date, created_at, updated_at 
              FROM events WHERE user_id = $1 AND event_date >= $2 AND event_date < $3`

	rows, err := p.pool.Query(ctx, query, userId, start, end)
	if err != nil {
		return nil, e.Wrap("storage.pg.GetEventsByCertainDay", err)
	}
	defer rows.Close()

	for rows.Next() {
		var event domain.Event
		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.EventDate,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			log.Printf("GetEventsByCertainDay: Scan error: %v", err)
			return nil, e.Wrap("storage.pg.GetEventsByCertainDay", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, e.Wrap("storage.pg.GetEventsByCertainDay", err)
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
