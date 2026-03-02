package tokenspec

import "fmt"

// TokenID represents a token declaration with an assigned numeric ID.
type TokenID struct {
	Name    string `json:"name"`
	ID      int    `json:"id"`
	Private bool   `json:"private"`
	Line    int    `json:"line"`
	Order   int    `json:"order"`
}

// AssignIDs assigns JavaCC-like token IDs:
// EOF is always 0, and declarations then increment from 1 by declaration order.
func AssignIDs(decls []TokenDecl) []TokenID {
	out := make([]TokenID, 0, len(decls)+1)
	out = append(out, TokenID{
		Name:  "EOF",
		ID:    0,
		Line:  0,
		Order: 0,
	})
	for _, d := range decls {
		out = append(out, TokenID{
			Name:    d.Name,
			ID:      d.Order,
			Private: d.Private,
			Line:    d.Line,
			Order:   d.Order,
		})
	}
	return out
}

// ToNameToID converts IDs to a name->id map, rejecting duplicate names.
func ToNameToID(ids []TokenID) (map[string]int, error) {
	m := make(map[string]int, len(ids))
	for _, id := range ids {
		if _, exists := m[id.Name]; exists {
			return nil, fmt.Errorf("duplicate token name: %s", id.Name)
		}
		m[id.Name] = id.ID
	}
	return m, nil
}
