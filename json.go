package main

// FilterValue is required by the list.Item interface.
func (i JSONResultItem) FilterValue() string { return i.Resource }

// Title is required by the list.Item interface.
func (i JSONResultItem) Title() string { return i.Resource }

// Description is required by the list.Item interface.
func (i JSONResultItem) Description() string {
	if i.Command != "" {
		return i.Command
	}
	return i.Message
}

// GetCategory returns the slice of results for a given category name.
func (jr *JSONResults) GetCategory(name string) []JSONResultItem {
	switch name {
	case "INFO":
		return jr.InfoResults
	case "OK":
		return jr.OkResults
	case "POTENTIAL_IMPORT":
		return jr.PotentialImportResults
	case "REGION_MISMATCH":
		return jr.RegionMismatchResults
	case "WARNING":
		return jr.WarningResults
	case "ERROR":
		return jr.ErrorResults
	case "DANGEROUS":
		return jr.DangerousResults
	default:
		return nil
	}
}
