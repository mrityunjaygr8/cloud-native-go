//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

type Events struct {
	Sequence  int32 `sql:"primary_key"`
	Key       string
	Value     string
	EventType int32
}
