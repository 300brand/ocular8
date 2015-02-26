package types

type LexisNexisPub struct {
	Name  string
	DPSI  string
	Track *bool
}

type LexisNexisFeed struct {
	Section string
	Track   *bool
}

type LexisNexisArticle struct {
	XML      []byte
	Filename string
	LNI      string
	DPSI     string
}
