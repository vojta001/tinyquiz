package schema

import (
	"entgo.io/ent"
)

// Provides a DBMS independent way of obtaining unique values
type CodesSequence struct {
	ent.Schema
}
