package data

import (
	"database/sql"
	"time"

	"github.com/dusktreader/the-hunt/internal/validator"
	"github.com/lib/pq"
)

type Company struct {
	ID			uint64		`json:"id"`
	CreatedAt	time.Time	`json:"created_at"`
	UpdatedAt	time.Time	`json:"updated_at"`
	Name		string		`json:"name"`
	URL			string		`json:"url,omitzero"`
	TechStack   []string	`json:"tech_stack,omitempty"`
}

func (c *Company) Validate(v *validator.Validator) {
	v.Check(c.Name != "", "name", "must be provided")
	v.Check(len(c.Name) <= 128, "name", "must not be more than 128 bytes")

	v.Check(validator.IsURL(c.URL), "url", "must be a valid URL")

	v.Check(c.TechStack != nil, "tech_stack", "must be provided")
	v.Check(len(c.TechStack) > 0, "tech_stack", "must not be empty")
	v.Check(len(c.TechStack) <= 5, "tech_stack", "must not be more than 5 items")
	v.Check(validator.Unique(c.TechStack), "tech_stack", "must not contain duplicate items")
}

type CompanyModel struct {
	DB *sql.DB
}

func (m CompanyModel) Insert(company *Company) error {
	query := `
		insert into companies (name, url, tech_stack)
		values ($1, $2, $3)
		returning id, created_at, updated_at
	`
	args := []any{company.Name, company.URL, pq.Array(company.TechStack)}
	return m.DB.QueryRow(query, args...).Scan(&company.ID, &company.CreatedAt, &company.UpdatedAt)
}

func (m CompanyModel) Get(id uint64) (*Company, error) {
	return nil, nil
}

func (m CompanyModel) Update(company *Company) error {
	return nil
}

func (m CompanyModel) Delete(id uint64) error {
	return nil
}
