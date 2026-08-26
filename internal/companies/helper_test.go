package companies

import (
	"testing"
)

func TestCompanyInputTrim(t *testing.T) {
	input := companyInput{
		Name:     "   Volvo  ",
		Website:  " https://volvo.com ",
		Industry: "\tAutomotive\n",
		Location: " Gothenburg ",
		Notes:    " some notes    ",
	}

	input.trim()

	if input.Name != "Volvo" {
		t.Errorf("expected Name to be trimmed, got %q", input.Name)
	}

	if input.Website != "https://volvo.com" {
		t.Errorf("expected Website to be trimmed, got %q", input.Website)
	}

	if input.Industry != "Automotive" {
		t.Errorf("expected Industry to be trimmed, got %q", input.Industry)
	}

	if input.Location != "Gothenburg" {
		t.Errorf("expected Location to be trimmed, got %q", input.Location)
	}

	if input.Notes != "some notes" {
		t.Errorf("expected Notes to be trimmed, got %q", input.Notes)
	}
}
