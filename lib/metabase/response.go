package metabase

import (
	"encoding/xml"
	"time"
)

// Structure taken from XSD found at:
// http://metabase.moreover.com/schema/articles10.xsd

type Response struct {
	XMLName          xml.Name  `xml:"response"`
	Status           string    `xml:"status,omitempty"`
	MessageCode      string    `xml:"messageCode,omitempty"`
	UserMessage      string    `xml:"userMessage,omitempty"`
	DeveloperMessage string    `xml:"developerMessage,omitempty"`
	Articles         []Article `xml:"articles>article,omitempty"`
}

type Article struct {
	SequenceId                  string                      `xml:"sequenceId,omitempty"`
	Id                          string                      `xml:"id"`
	Language                    string                      `xml:"language"`
	Title                       string                      `xml:"title"`
	SubTitle                    string                      `xml:"subTitle,omitempty"`
	Content                     string                      `xml:"content,omitempty"`
	ContentWithMarkup           string                      `xml:"contentWithMarkup,omitempty"`
	Tags                        []Tag                       `xml:"tags>tag,omitempty"`
	PublishedDate               string                      `xml:"publishedDate"`
	HarvestDate                 string                      `xml:"harvestDate"`
	EmbargoDate                 string                      `xml:"embargoDate,omitempty"`
	LicenseEndDate              string                      `xml:"licenseEndDate"`
	ContentLicenseEndDate       string                      `xml:"contentLicenseEndDate,omitempty"`
	Url                         string                      `xml:"url"`
	OriginalUrl                 string                      `xml:"originalUrl,omitempty"`
	CommentsUrl                 string                      `xml:"commentsUrl"`
	OutboundUrls                []OutboundUrl               `xml:"outboundUrls>outboundUrl,omitempty"`
	WordCount                   string                      `xml:"wordCount,omitempty"`
	DataFormat                  string                      `xml:"dataFormat"`
	Copyright                   string                      `xml:"copyright"`
	LoginStatus                 string                      `xml:"loginStatus"`
	DuplicateGroupId            string                      `xml:"duplicateGroupId"`
	ContentGroupIds             string                      `xml:"contentGroupIds,omitempty"`
	Harvest                     Harvest                     `xml:"harvest,omitempty"`
	Media                       Media                       `xml:"media"`
	PublishingPlatform          ArticlePublishingPlatform   `xml:"publishingPlatform"`
	AdultLanguage               string                      `xml:"adultLanguage"`
	Topics                      []Topic                     `xml:"topics>topic,omitempty"`
	Companies                   []Company                   `xml:"companies>company,omitempty"`
	Locations                   []ArticleLocation           `xml:"locations>location,omitempty"`
	Semantics                   Semantics                   `xml:"semantics,omitempty"`
	Sentiment                   Sentiment                   `xml:"sentiment,omitempty"`
	Print                       Print                       `xml:"print,omitempty"`
	Broadcast                   Broadcast                   `xml:"broadcast,omitempty"`
	Author                      Author                      `xml:"author"`
	Licenses                    []License                   `xml:"licenses>license,omitempty"`
	LinkedArticles              []LinkedArticle             `xml:"linkedArticles>linkedArticle,omitempty"`
	AdvertisingValueEquivalency AdvertisingValueEquivalency `xml:"advertisingValueEquivalency,omitempty"`
	Source                      Source                      `xml:"source"`
}

type Tag string

type OutboundUrl string

type Harvest struct {
	LegacyNewsFeedId           string        `xml:"legacyNewsFeedId,omitempty"`
	LegacyNewsName             string        `xml:"legacyNewsName,omitempty"`
	HarvestingMethods          string        `xml:"harvestingMethods,omitempty"`
	HarvesterServerId          string        `xml:"harvesterServerId,omitempty"`
	ProcessingTime             string        `xml:"processingTime,omitempty"`
	StreamReaderName           string        `xml:"streamReaderName,omitempty"`
	StreamReaderStartTime      string        `xml:"streamReaderStartTime,omitempty"`
	StreamReaderFileName       string        `xml:"streamReaderFileName,omitempty"`
	StreamReaderSearchTermName string        `xml:"streamReaderSearchTermName,omitempty"`
	CustomerTags               []CustomerTag `xml:"customerTags>customerTag,omitempty"`
}

type CustomerTag string

type Media struct {
	Audio []AudioOrVideo `xml:"audio,omitempty"`
	Image []Image        `xml:"image,omitempty"`
	Video []AudioOrVideo `xml:"video,omitempty"`
}

type AudioOrVideo struct {
	Url      string `xml:"url"`
	MimeType string `xml:"mimeType"`
	Caption  string `xml:"caption"`
	Duration string `xml:"duration"`
}

type Image struct {
	Url      string `xml:"url"`
	MimeType string `xml:"mimeType"`
	Caption  string `xml:"caption"`
}

type ArticlePublishingPlatform struct {
	ItemId            string      `xml:"itemId"`
	StatusId          string      `xml:"statusId"`
	ItemType          string      `xml:"itemType"`
	InReplyToUserId   string      `xml:"inReplyToUserId"`
	InReplyToStatusId string      `xml:"inReplyToStatusId"`
	UserMentions      UserMention `xml:"userMentions,omitempty"`
	TotalViews        string      `xml:"totalViews"`
}

type UserMention string

type Topic struct {
	Name  string `xml:"name,omitempty"`
	Group string `xml:"group,omitempty"`
}

type Company struct {
	Name         string `xml:"name"`
	Symbol       string `xml:"symbol"`
	Exchange     string `xml:"exchange"`
	Isin         string `xml:"isin"`
	Cusip        string `xml:"cusip,omitempty"`
	TitleCount   int    `xml:"titleCount"`
	ContentCount int    `xml:"contentCount"`
	Primary      bool   `xml:"primary"`
}

type ArticleLocation struct {
	Name       string  `xml:"name"`
	Type       string  `xml:"type"`
	Class      string  `xml:"class"`
	Mentions   string  `xml:"mentions"`
	Confidence string  `xml:"confidence"`
	Country    Country `xml:"country"`
	Region     string  `xml:"region"`
	Subregion  string  `xml:"subregion"`
	State      State   `xml:"state"`
	Latitude   string  `xml:"latitude"`
	Longitude  string  `xml:"longitude"`
	Provider   string  `xml:"provider"`
}

type Country struct {
	Confidence string `xml:"confidence,omitempty"`
	FipsCode   string `xml:"fipsCode,omitempty"`
	IsoCode    string `xml:"isoCode,omitempty"`
	Name       string `xml:"name,omitempty"`
}

type State struct {
	Confidence string `xml:"confidence,omitempty"`
	FipsCode   string `xml:"fipsCode,omitempty"`
	Name       string `xml:"name,omitempty"`
}

type Semantics struct {
	Events   []SemanticsItem `xml:"events>event,omitempty"`
	Entities []SemanticsItem `xml:"entities>entity,omitempty"`
}

type SemanticsItem struct {
	Properties []Property `xml:"properties>property,omitempty"`
}

type Property struct {
	Name  string `xml:"name,omitempty"`
	Value string `xml:"value,omitempty"`
}

type Sentiment struct {
	Score    string            `xml:"score,omitempty"`
	Entities []SentimentEntity `xml:"entities>entity,omitempty"`
}

type SentimentEntity struct {
	Type      string `xml:"type,omitempty"`
	Value     string `xml:"value,omitempty"`
	Mentions  string `xml:"mentions,omitempty"`
	Score     string `xml:"score,omitempty"`
	Evidence  string `xml:"evidence,omitempty"`
	Confident bool   `xml:"confident"`
}

type Print struct {
	Supplement         string `xml:"supplement"`
	PublicationEdition string `xml:"publicationEdition"`
	RegionalEdition    string `xml:"regionalEdition"`
	Section            string `xml:"section"`
	PageNumber         string `xml:"pageNumber"`
	PageCount          string `xml:"pageCount"`
	SizeCm             string `xml:"sizeCm"`
	SizePercentage     string `xml:"sizePercentage"`
	OriginLeft         string `xml:"originLeft"`
	OriginTop          string `xml:"originTop"`
	Width              string `xml:"width"`
	Height             string `xml:"height"`
	ByLine             string `xml:"byLine"`
	Photo              string `xml:"photo"`
}

type Broadcast struct {
	MarketName      string `xml:"marketName"`
	NationalNetwork string `xml:"nationalNetwork"`
	Title           string `xml:"title"`
	ProgramOrigin   string `xml:"programOrigin"`
	ProgramCategory string `xml:"programCategory"`
	Lines           []Line `xml:"lines>line,omitempty"`
}

type Line struct {
	Date string `xml:"date,omitempty"`
	Text string `xml:"text,omitempty"`
}

type Author struct {
	Name               string                   `xml:"name"`
	HomeUrl            string                   `xml:"homeUrl"`
	Email              string                   `xml:"email"`
	Description        string                   `xml:"description"`
	DateLastActive     string                   `xml:"dateLastActive"`
	PublishingPlatform AuthorPublishingPlatform `xml:"publishingPlatform"`
}

type AuthorPublishingPlatform struct {
	UserName       string `xml:"userName"`
	UserId         string `xml:"userId"`
	StatusesCount  string `xml:"statusesCount"`
	TotalViews     string `xml:"totalViews"`
	FollowingCount string `xml:"followingCount"`
	FollowersCount string `xml:"followersCount"`
	KloutScore     string `xml:"kloutScore"`
}

type License struct {
	Name string `xml:"name,omitempty"`
}

type LinkedArticle struct {
	Type      string `xml:"type,omitempty"`
	ArticleId string `xml:"articleId,omitempty"`
}

type AdvertisingValueEquivalency struct {
	Value    string `xml:"value,omitempty"`
	Currency string `xml:"currency,omitempty"`
}

type Source struct {
	Id                    string         `xml:"id,omitempty"`
	SyndicatorSourceId    string         `xml:"syndicatorSourceId,omitempty"`
	Name                  string         `xml:"name"`
	HomeUrl               string         `xml:"homeUrl"`
	Publisher             string         `xml:"publisher"`
	PrimaryLanguage       string         `xml:"primaryLanguage"`
	PrimaryMediaType      string         `xml:"primaryMediaType"`
	Category              string         `xml:"category"`
	EditorialRank         string         `xml:"editorialRank"`
	PublicationId         string         `xml:"publicationId,omitempty"`
	LicensorPublicationId string         `xml:"licensorPublicationId,omitempty"`
	ChannelCode           string         `xml:"channelCode,omitempty"`
	Location              SourceLocation `xml:"location"`
	Metrics               SourceMetric   `xml:"metrics,omitempty"`
	FeedSource            string         `xml:"feedSource,omitempty"`
	Feed                  Feed           `xml:"feed"`
}

type SourceLocation struct {
	Country     string `xml:"country"`
	CountryCode string `xml:"countryCode"`
	Region      string `xml:"region"`
	Subregion   string `xml:"subregion"`
	State       string `xml:"state"`
	County      string `xml:"county,omitempty"`
	ZipArea     string `xml:"zipArea"`
	ZipCode     string `xml:"zipCode"`
}

type SourceMetric struct {
	Mozscape     SourceMozData      `xml:"mozscape,omitempty"`
	ReachMetrics SourceReachMetrics `xml:"sourceReachMetrics,omitempty"`
}

type SourceMozData struct {
	MozRank         string `xml:"mozRank,omitempty"`
	PageAuthority   string `xml:"pageAuthority,omitempty"`
	DomainAuthority string `xml:"domainAuthority,omitempty"`
	ExternalLinks   string `xml:"externalLinks,omitempty"`
	Links           string `xml:"links,omitempty"`
}

type SourceReachMetrics struct {
	Reach []SourceReach `xml:"reach,omitempty"`
}

type SourceReach struct {
	Type      string `xml:"type,omitempty"`
	Value     string `xml:"value,omitempty"`
	Frequency string `xml:"frequency,omitempty"`
}

type Feed struct {
	Id                 int64    `xml:"id"`
	Name               string   `xml:"name"`
	Url                string   `xml:"url,omitempty"`
	MediaType          string   `xml:"mediaType"`
	PublishingPlatform string   `xml:"publishingPlatform"`
	IdFromPublisher    string   `xml:"idFromPublisher"`
	Generator          string   `xml:"generator"`
	Description        string   `xml:"description"`
	Tags               []Tag    `xml:"tags>tag,omitempty"`
	ImageUrl           string   `xml:"imageUrl"`
	Copyright          string   `xml:"copyright"`
	Language           string   `xml:"language"`
	DataFormat         string   `xml:"dataFormat"`
	Rank               Rank     `xml:"rank"`
	InWhiteList        string   `xml:"inWhiteList"`
	AutoTopics         []string `xml:"autoTopics>autoTopic,omitempty"`
	EditorialTopics    []string `xml:"editorialTopics>editorialTopic,omitempty"`
	Genre              string   `xml:"genre"`
}

type Rank struct {
	AutoRank         string `xml:"autoRank"`
	AutoRankOrder    string `xml:"autoRankOrder"`
	InboundLinkCount string `xml:"inboundLinkCount"`
}

func (r Response) NewSequenceId() string {
	if len(r.Articles) == 0 {
		return ""
	}
	i := len(r.Articles) - 1
	return r.Articles[i].SequenceId
}

func (a Article) Published() (t time.Time) {
	t, err := time.Parse(time.RFC3339, a.PublishedDate)
	if err != nil {
		t, _ = time.Parse(time.RFC3339, a.HarvestDate)
	}
	return
}

func (a Article) XML() (data []byte) {
	data, _ = xml.Marshal(a)
	return
}
