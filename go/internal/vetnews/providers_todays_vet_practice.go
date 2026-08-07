package vetnews

import "context"

const todaysVetPracticeRSS = "https://todaysveterinarypractice.com/feed/"

type todaysVetPracticeProvider struct {
	http *HTTPClient
}

func NewTodaysVetPracticeProvider(http *HTTPClient) Provider {
	return &todaysVetPracticeProvider{http: http}
}

func (p *todaysVetPracticeProvider) ID() string   { return "todays_vet_practice" }
func (p *todaysVetPracticeProvider) Name() string { return "Today's Veterinary Practice" }
func (p *todaysVetPracticeProvider) DefaultTags() []string {
	return []string{"pratique", "actualites", "en"}
}

func (p *todaysVetPracticeProvider) Fetch(ctx context.Context) ([]RawItem, error) {
	// category/news/ returns 404; site-wide WordPress feed is the stable source.
	body, err := p.http.Get(ctx, todaysVetPracticeRSS, "application/rss+xml, application/xml, text/xml")
	if err != nil {
		return nil, err
	}
	return ParseRSS(body, 20)
}
