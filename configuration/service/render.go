// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"bytes"
	"text/template"
)

func render(content string, data interface{}) ([]byte, error) {
	tpl, err := template.New("template").Parse(content)
	if err != nil {
		return nil, err
	}
	b := &bytes.Buffer{}
	if err := tpl.Execute(b, data); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
