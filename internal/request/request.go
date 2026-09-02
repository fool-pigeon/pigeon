package request

type Request interface {
	Run() (Result, error)
}

type TestResult struct {
	Name string
	// test success if Err==nil
	Err error
}

// As the View Result of [Request.Run]
type Result struct {
	Success bool
	ReqName string
	// this means the tag of status,such as "2xx","4xx","5xx","err","failed"
	Status string
	Msg    string
	// results of test
	Tests []TestResult
	// the raw content of response,e.g. [httpx.Response]
	Raw any
	// this is default nil if not get error in context
	Err error
}

type RequestMeta struct {
	Name     string `toml:"name"`
	Protocol string `toml:"protocol"`
}

const (
	// kinds of protocol
	P_HTTP = "http"
	P_GRPC = "grpc"
	P_WS   = "ws"
	P_GQL  = "gql"
)

var (
	Protocols = []string{P_HTTP, P_GRPC, P_WS, P_GQL}
)
