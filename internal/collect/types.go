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
	ExternalID  string
	Title       string
	Subtitle    string
	Description string
	Cover       string
	Poster      string
	Year        int32
	Region      string
	Language    string
	Duration    int32
	Rating      float64
	ReleaseDate string
	CategoryKey string // external category id/name for mapping
	Directors   []string
	Actors      []string
	Tags        []string
	Episodes    []Episode
}

// Page is one page of collected items.
type Page struct {
	Page      int
	PageCount int
	Total     int
	Items     []Item
}
