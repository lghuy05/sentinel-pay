package blacklist

const (
	ServiceName = "blacklist-service"
	DefaultPort = 8084
	InputTopic  = "transactions.enriched"
	OutputTopic = "fraud.blacklist"
)

type Entry struct {
	Type   string
	Value  string
	Active bool
}
