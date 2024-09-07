package uci

import (
	"encoding/json"
)

type scoreJSON struct {
	Mate  bool  `json:"mate"`
	Value int32 `json:"v"`
}

func (s Score) MarshalJSON() ([]byte, error) {
	return json.Marshal(scoreJSON{Mate: s.mate, Value: s.val})
}

func (s *Score) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var sj scoreJSON
	if err := json.Unmarshal(data, &sj); err != nil {
		return err
	}
	if sj.Mate {
		*s = ScoreMate(sj.Value)
	} else {
		*s = ScoreCentipawns(sj.Value)
	}
	return nil
}
