//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import (
	"time"
)

type Messages struct {
	ID        *int32     `sql:"primary_key" json:"id,omitempty"`
	ThreadID  *int32     `json:"thread_id,omitempty"`
	UserID    *int32     `json:"user_id,omitempty"`
	Body      *string    `json:"body,omitempty"`
	Type      *string    `json:"type,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}
