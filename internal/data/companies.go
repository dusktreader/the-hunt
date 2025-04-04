package data

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/dusktreader/the-hunt/internal/validator"
	"github.com/lib/pq"
)

type Company struct {
	ID			int64		`json:"id"`
	CreatedAt	time.Time	`json:"created_at"`
	UpdatedAt	time.Time	`json:"updated_at"`
	Name		string		`json:"name"`
	URL			string		`json:"url,omitzero"`
	TechStack   []string	`json:"tech_stack,omitempty"`
	Version		int64		`json:"version"`
}

type PartialCompany struct {
	Name		*string		`json:"name"`
	URL			*string		`json:"url"`
	TechStack   []string	`json:"tech_stack"`
}

type CompanyModel struct {
	DB *sql.DB
	CFG ModelConfig
}

var CompanySearchFields = NewSearchFields("name", "tech_stack")
var CompanySortFields = NewSortFields("id", "created_at", "updated_at", "name")
var CompanyInFields = NewInFields("tech_stack")

func (c *Company) Validate(v *validator.Validator) {
	v.Check(c.Name != "", "name", "must be provided")
	v.Check(len(c.Name) <= 128, "name", "must not be more than 128 bytes")

	v.Check(validator.IsURL(c.URL), "url", "must be a valid URL")

	v.Check(c.TechStack != nil, "tech_stack", "must be provided")
	v.Check(len(c.TechStack) > 0, "tech_stack", "must not be empty")
	v.Check(len(c.TechStack) <= 5, "tech_stack", "must not be more than 5 items")
	v.Check(validator.Unique(c.TechStack), "tech_stack", "must not contain duplicate items")
}

func (pc *PartialCompany) Validate(v *validator.Validator) {
	if pc.Name != nil {
		v.Check(*pc.Name != "", "name", "must be provided")
		v.Check(len(*pc.Name) <= 128, "name", "must not be more than 128 bytes")
	}

	if pc.URL != nil {
		v.Check(validator.IsURL(*pc.URL), "url", "must be a valid URL")
	}

	if pc.TechStack != nil {
		v.Check(len(pc.TechStack) > 0, "tech_stack", "must not be empty")
		v.Check(len(pc.TechStack) <= 5, "tech_stack", "must not be more than 5 items")
		v.Check(validator.Unique(pc.TechStack), "tech_stack", "must not contain duplicate items")
	}
}

func (m CompanyModel) GetVersion(id int64) (int64, error) {
	query := `
		select version
		from companies
		where id = $1
	`
	var version int64

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	return version, MapError(
		m.DB.QueryRowContext(ctx, query, id).Scan(&version),
		ErrorMap{sql.ErrNoRows: ErrRecordNotFound},
	)
}

func (m CompanyModel) Insert(company *Company) error {
	query := `
		insert into companies (name, url, tech_stack)
		values ($1, $2, $3)
		returning id, created_at, updated_at, version
	`
	args := []any{
		company.Name,
		company.URL,
		pq.Array(company.TechStack),
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	return MapError(
		m.DB.QueryRowContext(ctx, query, args...).Scan(
			&company.ID,
			&company.CreatedAt,
			&company.UpdatedAt,
			&company.Version,
		),
		ErrorMap{".*duplicate key.*": ErrDuplicateKey},
	)
}

func (m CompanyModel) GetOne(id int64) (*Company, error) {
	query := `
		select id, created_at, updated_at, name, url, tech_stack, version
		from companies
		where id = $1
	`
	var c Company

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	return &c, MapError(
		m.DB.QueryRowContext(ctx, query, id).Scan(
			&c.ID,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.Name,
			&c.URL,
			pq.Array(&c.TechStack),
			&c.Version,
		),
		ErrorMap{sql.ErrNoRows: ErrRecordNotFound},
	)
}

func (m CompanyModel) GetMany(f Filters) ([]*Company, *ListMetadata, error) {
	args := []any{}
	query_parts := []string{`
		select count(*) over (), id, created_at, updated_at, name, url, tech_stack, version
		from companies
	`}

	where_parts := []string{}

	if f.Search != nil {
		for k, v := range *f.Search {
			args = append(args, v)
			where_parts = append(where_parts, fmt.Sprintf("%s ~* $%d", k, len(args)))
			// Maybe try full text search down the road, but for now simple partial matching is what I want
			// Also consider using a gin index for partial matching
			// where_parts = append(where_parts, fmt.Sprintf("to_tsvector('simple', %s) @@ plainto_tsquery('simple', $%d)", k, i))
		}
	}

	if f.In != nil {
		for k, v := range *f.In {
			args = append(args, v)
			where_parts = append(where_parts, fmt.Sprintf("$%d = any(%s)", len(args), k))
		}
	}

	if len(where_parts) > 0 {
		query_parts = append(query_parts, "where", strings.Join(where_parts, " and "))
	}


	sort_parts := []string{}

	if f.Sort != nil {
		for k, v := range f.Sort.FromOldest() {
			sort_parts = append(sort_parts, fmt.Sprintf("%s %s", k, v))
		}
	}

	if len(sort_parts) > 0 {
		query_parts = append(query_parts, "order by", strings.Join(sort_parts, ", "))
	}

	if f.Page != nil && f.PageSize != nil {
		args = append(args, *f.PageSize, (*f.Page - 1) * *f.PageSize)
		query_parts = append(query_parts, fmt.Sprintf("limit $%d offset $%d", len(args) - 1, len(args)))
	}

	query := strings.Join(query_parts, " ")

	slog.Debug("Assembled GetMany query", "query", query, "args", args)

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var recordCount int
	companies := make([]*Company, 0, 10)
	for rows.Next() {
		var c Company
		err := rows.Scan(
			&recordCount,
			&c.ID,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.Name,
			&c.URL,
			pq.Array(&c.TechStack),
			&c.Version,
		)
		if err != nil {
			return nil, nil, err
		}
		companies = append(companies, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, nil, err
	}
	metadata := NewListMetadata(f, recordCount)
	return companies, &metadata, nil
}

func (m CompanyModel) Update(company *Company) error {
	query := `
		update companies
		set name = $1, url = $2, tech_stack = $3, updated_at = $4, version = version + 1
		where id = $5 and version = $6
		returning version
	`
	args := []any{
		company.Name,
		company.URL,
		pq.Array(company.TechStack),
		time.Now(),
		company.ID,
		company.Version,
	}

	return MapError(
		m.DB.QueryRow(query, args...).Scan(&company.Version),
		ErrorMap{
			sql.ErrNoRows: ErrEditConflict,
			".*duplicate key.*": ErrDuplicateKey,
		},
	)
}

func (m CompanyModel) PartialUpdate(
	id int64,
	version int64,
	partial *PartialCompany,
) (*Company, error) {
	query := `
		update companies
		set updated_at = $1, version = version + 1
	`
	args := []any{
		time.Now(),
	}

	i := 2
	if partial.Name != nil {
		query += fmt.Sprintf(", name = $%d", i)
		args = append(args, *partial.Name)
		i += 1
	}

	if partial.URL != nil {
		query += fmt.Sprintf(", url = $%d", i)
		args = append(args, *partial.URL)
		i += 1
	}

	if partial.TechStack != nil {
		query += fmt.Sprintf(", tech_stack = $%d", i)
		args = append(args, pq.Array(partial.TechStack))
		i += 1
	}

	query += fmt.Sprintf(`
		where id = $%d and version = $%d
		returning created_at, updated_at, name, url, tech_stack, version
	`, i, i+1)
	args = append(args, id, version)
	c := &Company{
		ID: id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	return c, MapError(
		m.DB.QueryRowContext(ctx, query, args...).Scan(
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.Name,
			&c.URL,
			pq.Array(&c.TechStack),
			&c.Version,
		),
		ErrorMap{sql.ErrNoRows: ErrEditConflict},
	)
}

func (m CompanyModel) Delete(id int64) error {
	query := `
		delete from companies
		where id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
