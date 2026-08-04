package contracts

const (
	TopicTransactionsRaw      = "transactions.raw"
	TopicTransactionsEnriched = "transactions.enriched"
	TopicFraudBlacklist       = "fraud.blacklist"
	TopicFraudRules           = "fraud.rules"
	TopicFraudML              = "fraud.ml"
	TopicFraudFinal           = "fraud.final"
)

var RequiredKafkaTopics = []string{
	TopicTransactionsRaw,
	TopicTransactionsEnriched,
	TopicFraudBlacklist,
	TopicFraudRules,
	TopicFraudML,
	TopicFraudFinal,
}
