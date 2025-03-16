package data

import (
	"time"

	"github.com/dusktreader/the-hunt/internal/validator"
)

type Company struct {
	ID			uint64		`json:"id"`
	CreatedAt	time.Time	`json:"created_at"`
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
