package httpx

const (
	AuthNone AuthType = iota
	AuthTypeBasic
	AuthTypeDigest
	AuthTypeToken
)

type (
	Retry struct {
		Count                int    `toml:"count"`
		WaitTime             string `toml:"wait_time"`
		MaxWaitTime          string `toml:"max_wait_time"`
		IsAllowNonIdempotent bool   `toml:"is_allow_non_idempotent"`
	}

	AuthType = int
	Auth     struct {
		Type AuthType `toml:"type" comment:"0:none,1:basic,2:digest(only-client),3:token,4:oauth(only-client)"`

		Token AuthToken `toml:"token" comment:"<header>: <scheme> <token>"`
		Creds AuthCreds `toml:"creds"`
	}
	AuthToken struct {
		Scheme string `toml:"scheme"`
		Token  string `toml:"token"`
		Header string `toml:"header"`
	}
	AuthCreds struct {
		Username string `toml:"username"`
		Password string `toml:"password"`
	}
	// TODO  type OAuth2.0 and OAuth1.0
)

type Setting struct {
	Timeout string `toml:"timeout" comment:"time>=0, e.g.:'1m30s' time_unit:ns|us|ms|s|m|h"`
	Retry   Retry  `toml:"retry"`

	Auth Auth `toml:"auth"`

	CookieString      string `toml:"cookies"`
	UnescapeQuery     bool   `toml:"unescape_query"`
	UnescapeJson      bool   `toml:"unescape_json"`
	ResponseBodyLimit int64  `toml:"response_body_limit" comment:"if limit<=0,it means no limit"`
}
