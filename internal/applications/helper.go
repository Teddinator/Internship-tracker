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
