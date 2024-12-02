package types

type OriginalChampionRates struct {
	Data map[string]map[string]struct {
		PlayRate float64 `json:"playRate"`
	} `json:"data"`
}

type RemappedChampionRates struct {
	Data map[string]map[string]struct {
		PlayRate float64 `json:"playRate"`
	} `json:"data"`
}
