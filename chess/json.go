package chess

import (
	"encoding/json"
	"fmt"
)

func (b RawBoard) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.FEN())
}

func (b *RawBoard) UnmarshalJSON(data []byte) error {
	var fen string
	if err := json.Unmarshal(data, &fen); err != nil {
		return err
	}
	res, err := RawBoardFromFEN(fen)
	if err != nil {
		return fmt.Errorf("raw board from fen: %w", err)
	}
	*b = res
	return nil
}

func (b *Board) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.FEN())
}

func (b *Board) UnmarshalJSON(data []byte) error {
	var fen string
	if err := json.Unmarshal(data, &fen); err != nil {
		return err
	}
	res, err := BoardFromFEN(fen)
	if err != nil {
		return fmt.Errorf("board from fen: %w", err)
	}
	*b = *res
	return nil
}

func (m UCIMove) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

func (m *UCIMove) UnmarshalJSON(data []byte) error {
	var uci string
	if err := json.Unmarshal(data, &uci); err != nil {
		return err
	}
	res, err := UCIMoveFromString(uci)
	if err != nil {
		return fmt.Errorf("uci move from string: %w", err)
	}
	*m = res
	return nil
}

func (c Color) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *Color) UnmarshalJSON(data []byte) error {
	var cs string
	if err := json.Unmarshal(data, &cs); err != nil {
		return err
	}
	res, err := ColorFromString(cs)
	if err != nil {
		return fmt.Errorf("color from string: %w", err)
	}
	*c = res
	return nil
}

func (c Coord) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *Coord) UnmarshalJSON(data []byte) error {
	var cs string
	if err := json.Unmarshal(data, &cs); err != nil {
		return err
	}
	res, err := CoordFromString(cs)
	if err != nil {
		return fmt.Errorf("coord from string: %w", err)
	}
	*c = res
	return nil
}

func (c MaybeCoord) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *MaybeCoord) UnmarshalJSON(data []byte) error {
	var cs string
	if err := json.Unmarshal(data, &cs); err != nil {
		return err
	}
	res, err := MaybeCoordFromString(cs)
	if err != nil {
		return fmt.Errorf("coord from string: %w", err)
	}
	*c = res
	return nil
}

func (c Cell) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *Cell) UnmarshalJSON(data []byte) error {
	var cs string
	if err := json.Unmarshal(data, &cs); err != nil {
		return err
	}
	res, err := CellFromString(cs)
	if err != nil {
		return fmt.Errorf("cell from string: %w", err)
	}
	*c = res
	return nil
}

func (r CastlingRights) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func (r *CastlingRights) UnmarshalJSON(data []byte) error {
	var rs string
	if err := json.Unmarshal(data, &rs); err != nil {
		return err
	}
	res, err := CastlingRightsFromString(rs)
	if err != nil {
		return fmt.Errorf("castling rights from string: %w", err)
	}
	*r = res
	return nil
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var ss string
	if err := json.Unmarshal(data, &ss); err != nil {
		return err
	}
	res, err := StatusFromString(ss)
	if err != nil {
		return fmt.Errorf("castling rights from string: %w", err)
	}
	*s = res
	return nil
}
