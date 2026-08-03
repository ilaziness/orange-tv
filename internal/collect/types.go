package collect

// Episode is a normalized playable episode.
type Episode struct {
	Number  int32
	Title   string
	URL     string
	Quality string
	Format  string
}

// Item is a normalized video item from any collect format.
type Item struct {
	ExternalID         string
	Title              string
	Subtitle           string
	Description        string
	Cover              string
	Poster             string
	Year               int32
	Region             string
	Language           string
	Duration           int32
	ReleaseDate        string
	ExternalCategoryID uint32 // external category type_id for mapping
	Directors          []string
	Actors             []string
	Tags               []string
	Episodes           []Episode
	VodTime            string // raw time string (vod_time / created_at) for time range filtering
	Remarks            string // vod_remarks, used to determine serial status
}

// RemoteCategory is one external category from a collect source.
// Generic across formats: Apple CMS class / Open API category.
type RemoteCategory struct {
	ID       uint32
	Name     string
	ParentID uint32
}

// Page is one page of collected items.
type Page struct {
	Page      int
	PageCount int
	Total     int
	Items     []Item
	Classes   []RemoteCategory
}

// ListPage is the parsed result of a list API response.
// It only extracts IDs and time list; detailed fields come from the detail API.
type ListPage struct {
	Page       int
	PageCount  int
	Total      int
	IDs        []uint32 // external ID list (vod_id for Apple CMS, video id for default)
	Times      []string // time list (vod_time / created_at), parallel to IDs, for time range filtering
	Categories []RemoteCategory
}
