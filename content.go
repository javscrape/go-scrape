package scrape

import "time"

// Genre ...
type Genre struct {
	URL     string
	Content string
}

// Sample ...
type Sample struct {
	Index int
	Thumb string
	Image string
	Title string
}

// Star ...
type Star struct {
	Image string
	Link  string
	Name  string   //english name
	Alias []string //other name(katakana,...)
}

// Content ...
type Content struct {
	From          string //where this
	Language      string
	Uncensored    bool
	ID            string
	Title         string
	OriginalTitle string
	Year          string
	ReleaseDate   time.Time
	Studio        string
	Director      string
	MovieSet      string
	Plot          string
	Genres        []*Genre
	Actors        []*Star
	Poster        string
	Thumb         string
	Sample        []*Sample
	Publisher     string
}
