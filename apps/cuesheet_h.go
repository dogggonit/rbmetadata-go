package apps

type CueTrackInfo struct {
	Title      string
	Performer  string
	SongWriter string
	// ms from start of track
	Offset uint64
}

type Cuesheet struct {
	Path       string
	File       string
	Title      string
	Performer  string
	SongWriter string

	TrackCount int
	Track      []CueTrackInfo

	CurrTrackIdx int
	CurrTrack    CueTrackInfo
}
