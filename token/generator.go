package token

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const (
	// DefaultClientID represents the Slack application client ID.
	DefaultClientID = "91899392594.382712253827"

	// DefaultClientSecret represents the Slack application client secret.
	DefaultClientSecret = "c7986be41b6ddb478041d1848dad5f6e"

	// DefaultScope represents the requested OAuth client permissions.
	DefaultScope = "client"

	// DefaultRedirectURI represents the OAuth callback URL hosted on GitHub Pages by the maintainer.
	DefaultRedirectURI = "https://jpbruinsslot.github.io/slack-term-auth/"

	// LegacyRedirectURI represents the deprecated redirect URL using the former maintainer handle.
	LegacyRedirectURI = "https://erroneousboat.github.io/slack-term-auth/"

	// DefaultFakeHost represents the dummy callback host utilized in manual extraction flows.
	DefaultFakeHost = "http://not.a.host"

	// SlackOAuthAuthorizeBaseURL is the OAuth v1 authorization endpoint.
	SlackOAuthAuthorizeBaseURL = "https://slack.com/oauth/authorize"

	// SlackOAuthAccessBaseURL is the OAuth v1 access token exchange endpoint.
	SlackOAuthAccessBaseURL = "https://slack.com/api/oauth.access"
)

// Generator builds URLs required for Slack OAuth token flows.
type Generator struct {
	ClientID     string
	ClientSecret string
	Scope        string
	RedirectURI  string
}

// NewGenerator returns an initialized Generator using maintainer defaults.
func NewGenerator() *Generator {
	return &Generator{
		ClientID:     DefaultClientID,
		ClientSecret: DefaultClientSecret,
		Scope:        DefaultScope,
		RedirectURI:  DefaultRedirectURI,
	}
}

// AuthorizeURL constructs the browser authorization URL.
// When redirectURI is non-empty, it overrides the default redirect URI.
func (g *Generator) AuthorizeURL(redirectURI string) string {
	targetRedirect := g.RedirectURI
	if redirectURI != "" {
		targetRedirect = redirectURI
	}

	params := url.Values{}
	params.Set("client_id", g.ClientID)
	params.Set("scope", g.Scope)
	params.Set("redirect_uri", targetRedirect)

	return fmt.Sprintf("%s?%s", SlackOAuthAuthorizeBaseURL, params.Encode())
}

// AccessURL constructs the token retrieval URL used to exchange an auth code.
func (g *Generator) AccessURL(code string) string {
	params := url.Values{}
	params.Set("client_id", g.ClientID)
	params.Set("client_secret", g.ClientSecret)
	params.Set("code", strings.TrimSpace(code))

	return fmt.Sprintf("%s?%s", SlackOAuthAccessBaseURL, params.Encode())
}

// CurlCommand produces the shell command to execute the access token request.
func (g *Generator) CurlCommand(code string) string {
	return fmt.Sprintf("curl -s \"%s\"", g.AccessURL(code))
}

// Method1AuthorizeURL returns the primary authorization URL directing to jpbruinsslot's callback page.
func Method1AuthorizeURL() string {
	return NewGenerator().AuthorizeURL(DefaultRedirectURI)
}

// Method2AuthorizeURL returns the fallback authorization URL redirecting to a dummy host.
func Method2AuthorizeURL() string {
	return NewGenerator().AuthorizeURL(DefaultFakeHost)
}

// TokenRetrievalURL returns the token exchange endpoint for the given authorization code.
func TokenRetrievalURL(code string) string {
	return NewGenerator().AccessURL(code)
}

// ExtractCode parses an authorization code from a callback URL or parameter string.
func ExtractCode(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", errors.New("empty authorization URL or code string")
	}

	if !strings.Contains(trimmed, "?") && !strings.Contains(trimmed, "=") && !strings.Contains(trimmed, "/") {
		return trimmed, nil
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("unable to parse URL: %v", err)
	}

	code := parsed.Query().Get("code")
	if code == "" {
		return "", errors.New("missing 'code' query parameter in callback URL")
	}

	return code, nil
}
