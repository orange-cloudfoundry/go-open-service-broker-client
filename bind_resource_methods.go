package v2

// IsNotEmpty returns true if either AppGUID or Route in the BindResource is not empty.
func (br *BindResource) IsNotEmpty() bool {
	return (br.AppGUID != nil && *br.AppGUID != "") || (br.Route != nil && *br.Route != "")
}
