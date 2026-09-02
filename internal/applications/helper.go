package applications

var validStatuses = map[string]struct{}{
	"applied":   {},
	"interview": {},
	"offer":     {},
	"rejected":  {},
	"withdrawn": {},
}

func isValidStatus(status string) bool {
	_, exists := validStatuses[status]
	return exists
}

func formattedApplicationResponse(a Application) applicationResponse {
	var appliedAt *string

	if a.AppliedAt != nil {
		formatted := a.AppliedAt.Format("2006-01-02")
		appliedAt = &formatted
	}

	return applicationResponse{
		ID:        a.ID,
		CompanyID: a.CompanyID,
		Company:   a.Company,
		Role:      a.Role,
		Status:    a.Status,
		AppliedAt: appliedAt,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}
