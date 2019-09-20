package graphql

import (
	"context"
	"database/sql"
	"fmt"
	archivist_server "github.com/cataclyst/archivist-server"
	"github.com/google/uuid"
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
	return &mutationResolver{r}
}

type documentResolver struct{ *Resolver }

func (r *documentResolver) Tags(ctx context.Context, obj *models.Document) ([]*models.Tag, error) {

	rows, err := r.Resolver.db.QueryContext(ctx,
		`select title, context from Tag where id in (
    		select tag_id from Document_Tag where document_id = ?)`, obj.ID)
	if err == sql.ErrNoRows {
		return []*models.Tag{}, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "Could not query tags for document")
	}

	var result []*models.Tag

	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.Title, &tag.Context); err != nil {
			return nil, errors.Wrap(err, "Could not scan row to Tag model")
		}
		result = append(result, &tag)
	}
	return result, nil
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) RecentDocuments(ctx context.Context) ([]*models.Document, error) {
	log.Printf("Getting recent documents...")
	sqlStmt := `select id, title, description, date from Document order by modified_at desc limit 20;`
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
	//time.Sleep(2000 * time.Millisecond)

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

	return &result, nil
}

func (r *queryResolver) Tags(ctx context.Context) ([]*models.Tag, error) {
	panic("not implemented")
}

type mutationResolver struct{ *Resolver }

const iso8601DateFormat = "2006-01-02"
const iso8601DateTimeFormat = "2006-01-02 15:04:05"

func (r *mutationResolver) CreateDocument(ctx context.Context, input archivist_server.DocumentInput) (*models.Document, error) {
	date, err := time.Parse(iso8601DateFormat, input.Date)
	if err != nil {
		return nil, errors.Wrap(err, "Could not parse document date")
	}
	date = date.UTC()

	log.Printf("Got these tags: %#v", input.Tags)

	documentID := uuid.New().String()
	currentTime := time.Now().Truncate(time.Millisecond).UTC()

	sqlStmt := `insert into Document (id, title, description, date, created_at, modified_at)
                values (?, ?, ?, ?, ?, ?)`
	_, err = r.Resolver.db.ExecContext(ctx, sqlStmt,
		documentID, input.Title, input.Description, date, currentTime, currentTime)
	if err != nil {
		return nil, errors.Wrap(err, "Could not insert document into database")
	}

	for _, tag := range input.Tags {
		row := r.Resolver.db.QueryRowContext(
			ctx, `select id from Tag where title = ? and context = ?`,
			tag.Title, tag.Context)
		var tagID string
		if err := row.Scan(&tagID); err == sql.ErrNoRows {
			log.Print("Tag does not yet exist. Creating it...")
			tagID = uuid.New().String()
			_, err = r.Resolver.db.ExecContext(
				ctx, `insert into Tag (id, title, context) values (?, ?, ?)`,
				tagID, tag.Title, tag.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not create new tag")
			}
		} else if err != nil {
			return nil, errors.Wrapf(err, "Could not check for existing tag %s:%s", tag.Title, tag.Context)
		}

		_, err := r.Resolver.db.ExecContext(ctx, `insert into Document_Tag (document_id, tag_id) values (?, ?)`,
			documentID, tagID)
		if err != nil {
			return nil, errors.Wrap(err, "Could not insert tag on document")
		}
	}

	log.Printf("Document created. New ID: %v", documentID)

	return &models.Document{
		ID: documentID,
		Title: input.Title,
		Description: input.Description,
		Date: date.Format(iso8601DateFormat),
		CreatedAt: currentTime.Format(iso8601DateTimeFormat),
		ModifiedAt: currentTime.Format(iso8601DateTimeFormat),
		TagIDs: nil,
	}, nil
}
