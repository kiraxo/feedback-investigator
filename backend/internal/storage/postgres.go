package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kiraxo/feedback-investigator/backend/graph/model"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(
	ctx context.Context,
	databaseURL string,
) (*PostgresRepository, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf(
			"create PostgreSQL connection pool: %w",
			err,
		)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf(
			"connect to PostgreSQL: %w",
			err,
		)
	}

	return &PostgresRepository{
		pool: pool,
	}, nil
}

func (r *PostgresRepository) Save(
	ctx context.Context,
	run *model.AgentRun,
) error {
	payload, err := json.Marshal(run)
	if err != nil {
		return fmt.Errorf("encode agent run: %w", err)
	}

	startedAt, err := time.Parse(
		time.RFC3339Nano,
		run.StartedAt,
	)
	if err != nil {
		return fmt.Errorf("parse startedAt: %w", err)
	}

	var completedAt any

	if run.CompletedAt != nil {
		parsedCompletedAt, parseErr := time.Parse(
			time.RFC3339Nano,
			*run.CompletedAt,
		)
		if parseErr != nil {
			return fmt.Errorf(
				"parse completedAt: %w",
				parseErr,
			)
		}

		completedAt = parsedCompletedAt
	}

	_, err = r.pool.Exec(
		ctx,
		`
		INSERT INTO agent_runs (
			id,
			status,
			mode,
			provider,
			feedback_count,
			started_at,
			completed_at,
			error_message,
			payload
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			mode = EXCLUDED.mode,
			provider = EXCLUDED.provider,
			feedback_count = EXCLUDED.feedback_count,
			started_at = EXCLUDED.started_at,
			completed_at = EXCLUDED.completed_at,
			error_message = EXCLUDED.error_message,
			payload = EXCLUDED.payload
		`,
		run.ID,
		run.Status.String(),
		run.Mode.String(),
		run.Provider.String(),
		run.FeedbackCount,
		startedAt,
		completedAt,
		run.ErrorMessage,
		payload,
	)
	if err != nil {
		return fmt.Errorf("save agent run: %w", err)
	}

	return nil
}

func (r *PostgresRepository) Get(
	ctx context.Context,
	id string,
) (*model.AgentRun, bool, error) {
	var payload []byte

	err := r.pool.QueryRow(
		ctx,
		`
		SELECT payload
		FROM agent_runs
		WHERE id = $1
		`,
		id,
	).Scan(&payload)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, fmt.Errorf(
			"get agent run: %w",
			err,
		)
	}

	run, err := decodeRun(payload)
	if err != nil {
		return nil, false, err
	}

	return run, true, nil
}

func (r *PostgresRepository) List(
	ctx context.Context,
	limit int,
) ([]*model.AgentRun, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.pool.Query(
		ctx,
		`
		SELECT payload
		FROM agent_runs
		ORDER BY created_at DESC
		LIMIT $1
		`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list agent runs: %w",
			err,
		)
	}
	defer rows.Close()

	runs := make([]*model.AgentRun, 0)

	for rows.Next() {
		var payload []byte

		if err := rows.Scan(&payload); err != nil {
			return nil, fmt.Errorf(
				"scan agent run: %w",
				err,
			)
		}

		run, err := decodeRun(payload)
		if err != nil {
			return nil, err
		}

		runs = append(runs, run)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate agent runs: %w",
			err,
		)
	}

	return runs, nil
}

func (r *PostgresRepository) ListTasks(
	ctx context.Context,
	runID *string,
) ([]*model.InvestigationTask, error) {
	if runID != nil {
		run, found, err := r.Get(ctx, *runID)
		if err != nil {
			return nil, err
		}

		if !found {
			return []*model.InvestigationTask{}, nil
		}

		return run.Tasks, nil
	}

	runs, err := r.List(ctx, 100)
	if err != nil {
		return nil, err
	}

	tasks := make([]*model.InvestigationTask, 0)

	for _, run := range runs {
		tasks = append(tasks, run.Tasks...)
	}

	return tasks, nil
}

func (r *PostgresRepository) Ping(
	ctx context.Context,
) error {
	return r.pool.Ping(ctx)
}

func (r *PostgresRepository) Close() {
	r.pool.Close()
}

func decodeRun(
	payload []byte,
) (*model.AgentRun, error) {
	var run model.AgentRun

	if err := json.Unmarshal(payload, &run); err != nil {
		return nil, fmt.Errorf(
			"decode stored agent run: %w",
			err,
		)
	}

	return &run, nil
}
