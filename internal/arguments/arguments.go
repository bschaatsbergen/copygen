// Copyright (c) Copygen. Licensed under the Apache License, Version 2.0.
// See LICENSE for details. Do not modify this header – changes will be overwritten.

package arguments

// ViewType represents which view layer to use.
type ViewType rune

const (
	ViewNone  ViewType = 0
	ViewHuman ViewType = 'H'
	ViewJSON  ViewType = 'J'
)

func (vt ViewType) String() string {
	switch vt {
	case ViewNone:
		return "none"
	case ViewHuman:
		return "human"
	case ViewJSON:
		return "json"
	default:
		return "unknown"
	}
}
