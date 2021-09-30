package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/deref/exo/internal/compstate/api"
	"github.com/deref/exo/internal/util/jsonutil"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	DB *sqlx.DB
}

func (sto *Store) SetState(ctx context.Context, input *api.SetStateInput) (*api.SetStateOutput, error) {
	if input.ComponentID == "" {
		return nil, fmt.Errorf("invalid component id: %q", input.ComponentID)
	}
	tagsJSON := "{}"
	if input.Tags != nil {
		tagsJSON = jsonutil.MustMarshalString(input.Tags)
	}
	tx, err := sto.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()
	row := sto.DB.QueryRowContext(ctx, `
		INSERT INTO component_state (
			component_id, version,
			type, content, tags, timestamp
		)
		VALUES (
			?, COALESCE((
					SELECT MAX(version) + 1
					FROM component_state
					WHERE component_id = ?
				), 1),
			?, ?, ?, ?
		)
		RETURNING version;
	`, input.ComponentID, input.ComponentID, input.Type, input.Content, tagsJSON, input.Timestamp)
	var output api.SetStateOutput
	if err := row.Scan(&output.Version); err != nil {
		return nil, fmt.Errorf("scanning: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing: %w", err)
	}
	return &output, nil
}

func (sto *Store) GetState(ctx context.Context, input *api.GetStateInput) (*api.GetStateOutput, error) {
	row := sto.DB.QueryRowContext(ctx, `
		SELECT
			component_id,
			version,
			type,
			content,
			tags,
			timestamp
		FROM component_state
		WHERE component_id = ?
		AND version = (
			SELECT MAX(version)
			FROM component_state
			WHERE component_id = ?
		)`, input.ComponentID, input.ComponentID)
	var output api.GetStateOutput
	var state api.State
	var tags string
	err := row.Scan(&state.ComponentID, &state.Version, &state.Type, &state.Content, &tags, &state.Timestamp)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	} else {
		if err := jsonutil.UnmarshalString(tags, &state.Tags); err != nil {
			return nil, fmt.Errorf("unmarshalling tags: %w", err)
		}
		output.State = &state
	}
	return &output, nil
}
