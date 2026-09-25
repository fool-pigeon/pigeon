package httpx

type (
	Retry struct {
		Count                int    `toml:"count" json:"count"`
		WaitTime             string `toml:"wait_time" json:"waitTime"`
		MaxWaitTime          string `toml:"max_wait_time" json:"maxWaitTime"`
		IsAllowNonIdempotent bool   `toml:"is_allow_non_idempotent" json:"isAllowNonIdempotent"`
	}

	Auth struct {
		Token  AuthToken `toml:"token,omitempty" comment:"<header>: <scheme> <val>" json:"token"`
		Baisc  AuthCreds `toml:"baisc,omitempty" json:"baisc"`
		Digest AuthCreds `toml:"digest,omitempty" json:"digest"`
	}
	AuthToken struct {
		Header string `toml:"header,omitempty" json:"header"`
		Scheme string `toml:"scheme,omitempty" json:"scheme"`
		Value  string `toml:"val,omitempty" json:"value"`
	}
	AuthCreds struct {
		Username string `toml:"username" json:"username"`
		Password string `toml:"password" json:"password"`
	}
	// TODO  type OAuth2.0 and OAuth1.0
)

type Setting struct {
	Timeout string `toml:"timeout" comment:"time>=0, e.g.:'1m30s' time_unit:ns|us|ms|s|m|h" json:"timeout"`
	Retry   Retry  `toml:"retry" json:"retry"`

	Auth Auth `toml:"auth" comment:"token,basic,digest(only-client)" json:"auth"`

	CookieString      string `toml:"cookies" json:"cookieString"`
	UnescapeQuery     bool   `toml:"unescape_query" json:"unescapeQuery"`
	UnescapeJson      bool   `toml:"unescape_json" json:"unescapeJson"`
	ResponseBodyLimit int64  `toml:"response_body_limit" comment:"if limit<=0,it means no limit" json:"responseBodyLimit"`
}
