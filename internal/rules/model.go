package rules

const (
	ServiceName = "rule-engine"
	DefaultPort = 8083
	InputTopic  = "fraud.blacklist"
	OutputTopic = "fraud.rules"
)

type Rule struct {
	Name      string
	Type      string
	Threshold float64
	Weight    float64
	Active    bool
}
