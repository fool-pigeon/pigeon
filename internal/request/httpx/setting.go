package httpx

type (
	Retry struct {
		Count                int    `toml:"count"`
		WaitTime             string `toml:"wait_time"`
		MaxWaitTime          string `toml:"max_wait_time"`
		IsAllowNonIdempotent bool   `toml:"is_allow_non_idempotent"`
	}

	Auth struct {
		Token  AuthToken `toml:"token,omitempty" comment:"<header>: <scheme> <val>"`
		Baisc  AuthCreds `toml:"baisc,omitempty"`
		Digest AuthCreds `toml:"digest,omitempty"`
	}
	AuthToken struct {
		Header string `toml:"header,omitempty"`
		Scheme string `toml:"scheme,omitempty"`
		Value  string `toml:"val,omitempty"`
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

	Auth Auth `toml:"auth" comment:"token,basic,digest(only-client)"`

	CookieString      string `toml:"cookies"`
	UnescapeQuery     bool   `toml:"unescape_query"`
	UnescapeJson      bool   `toml:"unescape_json"`
	ResponseBodyLimit int64  `toml:"response_body_limit" comment:"if limit<=0,it means no limit"`
}
