package graphql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"

	"github.com/cataclyst/archivist-server/models"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct {
	db *sql.DB
}

func (r *Resolver) Document() DocumentResolver {
	return &documentResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}
func (r *Resolver) Mutation() MutationResolver {
	panic("implement me")
}

type documentResolver struct{ *Resolver }

func (r *documentResolver) Tags(ctx context.Context, obj *models.Document) ([]*models.Tag, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) RecentDocuments(ctx context.Context) ([]*models.Document, error) {
	log.Printf("Getting recent documents...")
	sqlStmt := `select id, title, description, date from Document order by modified_at limit 20;`
	var result []*models.Document
	rows, err := r.Resolver.db.QueryContext(ctx, sqlStmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}
	defer rows.Close()

	for rows.Next() {
		var document models.Document
		err := rows.Scan(
			&document.ID,
			&document.Title,
			&document.Description,
			&document.Date)
		if err != nil {
			return nil, errors.Wrap(err, "Could not scan database row to models.Document")
		}
		result = append(result, &document)
	}

	// TODO remove - just to simulate latency
	time.Sleep(2000 * time.Millisecond)

	return result, nil
}

func (r *queryResolver) Document(ctx context.Context, id string) (*models.Document, error) {
	log.Printf("Getting document for ID %s...", id)
	sqlStmt := `select id, title, description, date from Document where id = ?;`
	row := r.Resolver.db.QueryRowContext(ctx, sqlStmt, id)

	var result models.Document
	err := row.Scan(
		&result.ID,
		&result.Title,
		&result.Description,
		&result.Date)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no document found with ID %s", id)
		}
		return nil, errors.Wrap(err, "Could not scan database row to models.Document")
	}

	log.Printf("the result: %v", result)

	return &result, nil
}

func (r *queryResolver) Tags(ctx context.Context) ([]*models.Tag, error) {
	panic("not implemented")
}
