package followups

func toFollowUpResponse(f FollowUp) FollowUpResponse {
	return FollowUpResponse{
		ID:            f.ID,
		ApplicationID: f.ApplicationID,
		DueDate:       f.DueDate.Format("2006-01-02"),
		Message:       f.Message,
		CompletedAt:   f.CompletedAt,
		CreatedAt:     f.CreatedAt,
		UpdatedAt:     f.UpdatedAt,
	}
}
