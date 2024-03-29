package graphql

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/cataclyst/archivist-server/models"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

const documentFileDir = "./docs/"

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
		// TODO really fatal here?
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

func (r *queryResolver) Search(ctx context.Context, term string) ([]*models.Document, error) {
	log.Printf("Searching for documents with %s...", term)

	fuzzyTerm := "%" + term + "%"
	searchStatement := `
		select id, title, description, date from Document D
		where D.title       like ?
        or    D.description like ?
        or    D.id in (
			select DT.document_id
            from   Document_Tag DT
			where  DT.tag_id in (select T.id from Tag T where T.title like ? or T.context like ?)
		)`

	var result []*models.Document
	rows, err := r.Resolver.db.QueryContext(ctx, searchStatement, fuzzyTerm, fuzzyTerm, fuzzyTerm, fuzzyTerm)
	if err != nil {
		// TODO really fatal here?
		log.Fatalf("%q: %s\n", err, searchStatement)
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

	return result, nil
}

type mutationResolver struct{ *Resolver }

const iso8601DateFormat = "2006-01-02"
const iso8601DateTimeFormat = "2006-01-02 15:04:05"

func (r *mutationResolver) CreateOrUpdateDocument(ctx context.Context, input models.DocumentInput) (*models.Document, error) {
	date, err := time.Parse(iso8601DateFormat, input.Date)
	if err != nil {
		return nil, errors.Wrap(err, "Could not parse document date")
	}
	date = date.UTC()

	log.Printf("Got these tags: %#v", input.Tags)

	documentID := input.ID
	if documentID == "" {
		documentID = uuid.New().String()
	}
	currentTime := time.Now().Truncate(time.Millisecond).UTC()

	var originalFileName string
	var mimeType string
	if input.DocumentData != nil {
		originalFileName = input.DocumentData.FileName
		mimeType = input.DocumentData.MimeType
	}

	// Try updating the document first, in case it already exists:
	updateStatement := `
		update Document
		set title = ?, description = ?, date = ?, document_file_name = ?, document_mime_type = ?, modified_at = ?
		where id = ?`

	execInfo, err := r.Resolver.db.ExecContext(ctx, updateStatement,
		input.Title, input.Description, date, originalFileName, mimeType, currentTime, documentID)
	if err != nil {
		return nil, errors.Wrap(err, "Could not (try to) update document")
	}
	rowsUpdated, err := execInfo.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, "Could not resolve number of updated documents")
	}

	// Document does not exist yet -- create it:
	if rowsUpdated == 0 {
		insertStatement := `insert into Document (id, title, description, date, document_mime_type, created_at, modified_at)
                values (?, ?, ?, ?, ?, ?, ?)`
		_, err = r.Resolver.db.ExecContext(ctx, insertStatement,
			documentID, input.Title, input.Description, date, mimeType, currentTime, currentTime)
		if err != nil {
			return nil, errors.Wrap(err, "Could not insert document into database")
		}
	}

	_, err = r.Resolver.db.ExecContext(ctx, `delete from Document_Tag where document_id = ?`, documentID)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not delete tags from document %s", documentID)
	}

	for _, tag := range input.Tags {
		tagID, err := r.ensureTagExists(ctx, tag)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not ensure tag '%s' exists", tag.Title)
		}
		_, err = r.Resolver.db.ExecContext(ctx,
			`insert into Document_Tag (document_id, tag_id) values (?, ?)`,
			documentID, tagID)
		if err != nil {
			return nil, errors.Wrap(err, "Could not insert tag on document")
		}
	}

	// Store the binary file data for the document.
	// The file name of the resulting file is "<uuid>-<original file name>".
	if input.DocumentData != nil {

		inputDocumentData := *input.DocumentData
		documentData, err := base64.StdEncoding.DecodeString(inputDocumentData.BinaryDataBase64)
		if err != nil {
			return nil, errors.Wrap(err, "Could not decode document data - is it properly Base64 encoded?")
		}

		filename := documentFileDir + documentID + "-" + originalFileName

		err = os.MkdirAll(documentFileDir, 0644)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not create document file directory: %s", documentFileDir)
		}

		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not open file descriptor for writing: %s", filename)
		}
		defer file.Close()

		_, err = file.Write(documentData)
		if err != nil {
			return nil, errors.Wrap(err, "Could not write document data to file")
		}
	}

	return &models.Document{
		ID:          documentID,
		Title:       input.Title,
		Description: input.Description,
		Date:        date.Format(iso8601DateFormat),
		CreatedAt:   currentTime.Format(iso8601DateTimeFormat),
		ModifiedAt:  currentTime.Format(iso8601DateTimeFormat),
		TagIDs:      nil,
	}, nil
}

func (r *mutationResolver) ensureTagExists(ctx context.Context, tag *models.TagInput) (string, error) {
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
			return "", errors.Wrapf(err, "Could not create new tag")
		}
	} else if err != nil {
		return "", errors.Wrapf(err, "Could not check for existing tag %s:%s", tag.Title, tag.Context)
	}
	return tagID, nil
}
