package repositories

type Stats struct {
	TotalAliases   int64 `json:"total_aliases"`
	TotalForwarded int64 `json:"total_forwarded"`
	TotalBlocked   int64 `json:"total_blocked"`
}
