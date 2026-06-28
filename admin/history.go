package admin

// HistoryPage stores render-ready object history.
type HistoryPage struct {
	ContentType string
	ObjectID    string
	Entries     []AdminLogEntry
}

// BuildHistoryPage filters log entries for one object.
func BuildHistoryPage(store *MemoryLogStore, contentType, objectID string) HistoryPage {
	page := HistoryPage{ContentType: contentType, ObjectID: objectID}
	if store == nil {
		return page
	}
	for _, entry := range store.Entries {
		if entry.ContentType == contentType && entry.ObjectID == objectID {
			page.Entries = append(page.Entries, entry)
		}
	}
	return page
}
