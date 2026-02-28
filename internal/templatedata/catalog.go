package templatedata

// GetAllTemplates returns all template categories (builtin + hackathon + solopreneur).
func GetAllTemplates() []TemplateCategory {
	all := make([]TemplateCategory, 0)
	all = append(all, GetBuiltinTemplates()...)
	all = append(all, GetHackathonCategory())
	all = append(all, GetSolopreneurCategory())
	return all
}

// GetTemplateByID finds a template by ID, returns nil if not found.
func GetTemplateByID(id string) *Template {
	for _, cat := range GetAllTemplates() {
		for i := range cat.Templates {
			if cat.Templates[i].ID == id {
				t := cat.Templates[i]
				return &t
			}
		}
	}
	return nil
}
