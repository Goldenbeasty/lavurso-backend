//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

type Years struct {
	ID          *int32  `sql:"primary_key" json:"id,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	Courses     *int32  `json:"courses,omitempty"`
	Current     *bool   `json:"current,omitempty"`
}
