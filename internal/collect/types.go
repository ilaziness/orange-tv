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
	ExternalCategoryID int64 // external category type_id for mapping
	Directors          []string
	Actors             []string
	Tags               []string
	Episodes           []Episode
	VodTime            string // raw vod_time string for time range filtering
	Remarks            string // vod_remarks, used to determine serial status
}

// AppleCMSClass is one category from Apple CMS class field.
type AppleCMSClass struct {
	TypeID   int64  `json:"type_id"`
	TypeName string `json:"type_name"`
	TypePID  int64  `json:"type_pid"`
}

// Page is one page of collected items.
type Page struct {
	Page      int
	PageCount int
	Total     int
	Items     []Item
	Classes   []AppleCMSClass
}

// ListPage is the parsed result of an Apple CMS list API response.
// It only extracts vod_id list and class info; detailed fields come from the detail API.
type ListPage struct {
	Page      int
	PageCount int
	Total     int
	VodIDs    []int64  // vod_id list (integer IDs)
	VodTimes  []string // vod_time list, parallel to VodIDs, for time range filtering
	Classes   []AppleCMSClass
}
