package vetnews

import "context"

const vetEvidenceRSS = "https://veterinaryevidence.org/index.php/ve/gateway/plugin/WebFeedGatewayPlugin/rss2"

type veterinaryEvidenceProvider struct {
	http *HTTPClient
}

func NewVeterinaryEvidenceProvider(http *HTTPClient) Provider {
	return &veterinaryEvidenceProvider{http: http}
}

func (p *veterinaryEvidenceProvider) ID() string   { return "veterinary_evidence" }
func (p *veterinaryEvidenceProvider) Name() string { return "Veterinary Evidence (RCVS Knowledge)" }
func (p *veterinaryEvidenceProvider) DefaultTags() []string {
	return []string{"ebm", "science", "en"}
}

func (p *veterinaryEvidenceProvider) Fetch(ctx context.Context) ([]RawItem, error) {
	body, err := p.http.Get(ctx, vetEvidenceRSS, "application/rss+xml, application/xml, text/xml")
	if err != nil {
		return nil, err
	}
	return ParseRSS(body, 20)
}
