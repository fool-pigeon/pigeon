package request

type Request interface {
	Run() Result
	ReqName() string
	ReqProtocol() string
}

type TestRes struct {
	Name   string
	Passed bool
	Err    error
}

// As the View Result of [Request.Run]
type Result struct {
	Success bool
	ReqName string
	// this means the tag of status,such as "2xx","4xx","5xx","err","failed"
	Status string
	// results of test
	Tests []TestRes
	// the raw content of response,e.g. [httpx.Response]
	Raw any
	// this is default nil if not get error in context
	Err error
}

func (result *Result) SetSuccess(Success bool) *Result {
	result.Success = Success
	return result
}

func (result *Result) SetReqName(ReqName string) *Result {
	result.ReqName = ReqName
	return result
}

func (result *Result) SetStatus(Status string) *Result {
	result.Status = Status
	return result
}

func (result *Result) SetTests(Tests []TestRes) *Result {
	result.Tests = Tests
	return result
}

func (result *Result) SetRaw(Raw any) *Result {
	result.Raw = Raw
	return result
}

func (result *Result) SetErr(Err error) *Result {
	result.Err = Err
	return result
}

func (r *Result) AppendTestRes(tRes TestRes) {
	r.Tests = append(r.Tests, tRes)
}
func NewResult() Result {
	return Result{}
}
func ErrResult(reqName string, err error) Result {
	return Result{ReqName: reqName, Err: err}
}

type RequestMeta struct {
	Name     string `toml:"name"`
	Protocol string `toml:"protocol"`
}

// ReqName implements [Request].
func (r *RequestMeta) ReqName() string {
	return r.Name
}

// ReqProtocol implements [Request].
func (r *RequestMeta) ReqProtocol() string {
	return r.Protocol
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
