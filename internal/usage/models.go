package usage

// Snapshot 暴露给前端的统一快照结构
type Snapshot struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	ShortName     string       `json:"shortName"`
	Color         string       `json:"color"`
	Emoji         string       `json:"emoji"`
	Primary       *QuotaWindow `json:"primary"`
	Secondary     *QuotaWindow `json:"secondary"`
	Token         TokenStats   `json:"token"`
	Credits       *CreditsInfo `json:"credits"`
	Status        string       `json:"status"`
	Source        string       `json:"source"`
	QuotaSource   string       `json:"quotaSource"` // local | online | none
	LastRequestAt int64        `json:"lastRequestAt"`
	Ts            int64        `json:"ts"`
	Error         string       `json:"error"`
}

type QuotaWindow struct {
	UsedPct       float64 `json:"usedPct"`
	RemainingPct  float64 `json:"remainingPct"`
	WindowMinutes int     `json:"windowMinutes"`
	ResetsAt      *int64  `json:"resetsAt"`
}

type TokenStats struct {
	Today        int64   `json:"today"`
	Near7d       int64   `json:"near7d"`
	Total        int64   `json:"total"`
	CurrentTask  int64   `json:"currentTask"`
	LastTurn     int64   `json:"lastTurn"`
	CacheHitRate float64 `json:"cacheHitRate"`
}

type CreditsInfo struct {
	HasCredits bool   `json:"hasCredits"`
	Unlimited  bool   `json:"unlimited"`
	Balance    string `json:"balance"`
}
