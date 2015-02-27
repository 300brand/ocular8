package types

type LexisNexisPub struct {
	Name  string
	Id    string // Could be ReportId or JournalCode or DSPI
	Track *bool
}

type LexisNexisFeed struct {
	Section string
	Track   *bool
}

type LexisNexisArticle struct {
	XML         []byte
	Filename    string
	LNI         string
	DPSI        string
	ReportNo    string
	JournalCode string
}
