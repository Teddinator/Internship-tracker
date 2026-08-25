package companies

import (
	"strings"
)

func (i *companyInput) trim() {
	i.Name = strings.TrimSpace(i.Name)
	i.Website = strings.TrimSpace(i.Website)
	i.Industry = strings.TrimSpace(i.Industry)
	i.Location = strings.TrimSpace(i.Location)
	i.Notes = strings.TrimSpace(i.Notes)
}
