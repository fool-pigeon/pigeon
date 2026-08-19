package request

type Result interface{}
type Request interface {
	Run() (Result, error)
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
