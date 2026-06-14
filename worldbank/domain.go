package worldbank

import (
	"context"
	"strings"
	"unicode"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes the World Bank Open Data API as a kit Domain: a driver that
// a multi-domain host (ant) enables with a single blank import,
//
//	import _ "github.com/tamnd/worldbank-cli/worldbank"
//
// exactly as a database/sql program enables a driver with `import _
// "github.com/lib/pq"`. The init below registers it; the host then dereferences
// worldbank:// URIs by routing to the operations Register installs.
func init() { kit.Register(Domain{}) }

// Domain is the World Bank driver. It carries no state; the per-run client is
// built by the factory Register hands kit.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme:  "worldbank",
		Aliases: []string{"wb"},
		Hosts:   []string{Host, "data.worldbank.org"},
		Identity: kit.Identity{
			Binary: "worldbank",
			Short:  "Read World Bank Open Data",
			Long: `Read World Bank Open Data

worldbank reads public World Bank data over plain HTTPS, shapes it into
clean records, and prints output that pipes into the rest of your tools.
No API key, nothing to run alongside it.`,
			Site: "data.worldbank.org",
			Repo: "https://github.com/tamnd/worldbank-cli",
		},
	}
}

// Register installs the client factory and every World Bank operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{
		Name:    "countries",
		Group:   "data",
		Summary: "List World Bank countries",
	}, listCountries)

	kit.Handle(app, kit.OpMeta{
		Name:    "indicators",
		Group:   "data",
		Summary: "List World Development Indicators",
	}, listIndicators)

	kit.Handle(app, kit.OpMeta{
		Name:    "data",
		Group:   "data",
		Summary: "Fetch indicator data for a country",
		Args:    []kit.Arg{{Name: "country", Help: "country code(s) e.g. US or US;CN;DE"}},
	}, getData)

	kit.Handle(app, kit.OpMeta{
		Name:    "topics",
		Group:   "data",
		Summary: "List World Bank thematic topics",
	}, listTopics)
}

// newClient builds the World Bank client from the host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := NewClient()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.HTTP.Timeout = cfg.Timeout
	}
	return c, nil
}

// --- inputs ---

type countriesInput struct {
	Region string  `kit:"flag" help:"filter by region code e.g. EAS,LCN,SSF"`
	Limit  int     `kit:"flag,inherit" help:"max results" default:"20"`
	Client *Client `kit:"inject"`
}

type indicatorsInput struct {
	Limit  int     `kit:"flag,inherit" help:"max results" default:"20"`
	Client *Client `kit:"inject"`
}

type dataInput struct {
	Country   string  `kit:"arg" help:"country code(s) e.g. US or US;CN;DE"`
	Indicator string  `kit:"flag" help:"indicator ID" default:"NY.GDP.MKTP.CD"`
	MRV       int     `kit:"flag" help:"most recent N values" default:"5"`
	Client    *Client `kit:"inject"`
}

type topicsInput struct {
	Client *Client `kit:"inject"`
}

// --- handlers ---

func listCountries(ctx context.Context, in countriesInput, emit func(*Country) error) error {
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	countries, err := in.Client.ListCountries(ctx, in.Region, limit)
	if err != nil {
		return mapErr(err)
	}
	for i := range countries {
		if err := emit(&countries[i]); err != nil {
			return err
		}
	}
	return nil
}

func listIndicators(ctx context.Context, in indicatorsInput, emit func(*Indicator) error) error {
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	indicators, err := in.Client.ListIndicators(ctx, limit)
	if err != nil {
		return mapErr(err)
	}
	for i := range indicators {
		if err := emit(&indicators[i]); err != nil {
			return err
		}
	}
	return nil
}

func getData(ctx context.Context, in dataInput, emit func(*DataPoint) error) error {
	indicator := in.Indicator
	if indicator == "" {
		indicator = "NY.GDP.MKTP.CD"
	}
	mrv := in.MRV
	if mrv <= 0 {
		mrv = 5
	}
	points, err := in.Client.GetData(ctx, in.Country, indicator, mrv)
	if err != nil {
		return mapErr(err)
	}
	for i := range points {
		if err := emit(&points[i]); err != nil {
			return err
		}
	}
	return nil
}

func listTopics(ctx context.Context, in topicsInput, emit func(*Topic) error) error {
	topics, err := in.Client.ListTopics(ctx)
	if err != nil {
		return mapErr(err)
	}
	for i := range topics {
		if err := emit(&topics[i]); err != nil {
			return err
		}
	}
	return nil
}

// --- Resolver ---

// Classify turns any accepted input into (uriType, id):
//   - 2-letter uppercase or country URL → ("country", id)
//   - indicator pattern (contains dots) → ("indicator", id)
//   - numeric string → ("topic", id)
//   - otherwise → ("country", id) as best guess
func (Domain) Classify(input string) (uriType, id string, err error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", errs.Usage("empty input")
	}
	// strip https://data.worldbank.org/ URLs
	if strings.HasPrefix(input, "http") {
		if strings.Contains(input, "/country/") {
			parts := strings.SplitN(input, "/country/", 2)
			id = strings.Trim(parts[1], "/")
			return "country", strings.ToUpper(id), nil
		}
		if strings.Contains(input, "/indicator/") {
			parts := strings.SplitN(input, "/indicator/", 2)
			id = strings.Trim(parts[1], "/")
			return "indicator", id, nil
		}
		return "", "", errs.Usage("unrecognized worldbank URL: %q", input)
	}
	// numeric → topic
	if isNumeric(input) {
		return "topic", input, nil
	}
	// contains dots → indicator (e.g. NY.GDP.MKTP.CD)
	if strings.Contains(input, ".") {
		return "indicator", input, nil
	}
	// 2-letter uppercase → country ISO2
	if len(input) == 2 && isUpperAlpha(input) {
		return "country", strings.ToUpper(input), nil
	}
	// default: treat as country
	return "country", strings.ToUpper(input), nil
}

// Locate returns the data.worldbank.org URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "country":
		return "https://data.worldbank.org/country/" + id, nil
	case "indicator":
		return "https://data.worldbank.org/indicator/" + id, nil
	case "topic":
		return "https://data.worldbank.org/topic/" + id, nil
	default:
		return "", errs.Usage("worldbank has no resource type %q", uriType)
	}
}

// --- helpers ---

func isNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

func isUpperAlpha(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "not found") {
		return errs.NotFound("%s", err.Error())
	}
	return err
}
