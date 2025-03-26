// Copyright (c) Copygen. Licensed under the Apache License, Version 2.0.
// See LICENSE for details. Do not modify this header – changes will be overwritten.

package view

import "io"

type View struct {
	Stream *Stream
}

type Stream struct {
	Writer io.Writer
}

func (s *Stream) Write(p []byte) (n int, err error) {
	return s.Writer.Write(p)
}

func NewView(w io.Writer) *View {
	return &View{
		Stream: &Stream{
			Writer: w,
		},
	}
}
