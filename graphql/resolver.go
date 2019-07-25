package graphql

import (
	"context"
	"database/sql"
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
func (r *queryResolver) Tags(ctx context.Context) ([]*models.Tag, error) {
	panic("not implemented")
}
