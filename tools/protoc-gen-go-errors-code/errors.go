// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package main

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	"github.com/superproj/onex/tools/protoc-gen-go-errors-code/errors"
)

var enCases = cases.Title(language.AmericanEnglish, cases.NoLower)

// generateFile generates a _code.md file containing kratos errors definitions.
func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Enums) == 0 {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + "_code.md"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	generateFileContent(gen, file, g)
	return g
}

// generateFileContent generates the kratos errors definitions, excluding the package statement.
func generateFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	if len(file.Enums) == 0 {
		return
	}

	index := 0
	for _, enum := range file.Enums {
		if !genErrorsReason(gen, file, g, enum) {
			index++
		}
	}
	// If all enums do not contain 'errors.code', the current file is skipped
	if index == 0 {
		g.Skip()
	}
}

func genErrorsReason(_ *protogen.Plugin, _ *protogen.File, g *protogen.GeneratedFile, enum *protogen.Enum) bool {
	defaultCode := proto.GetExtension(enum.Desc.Options(), errors.E_DefaultCode)
	code := 0
	if ok := defaultCode.(int32); ok != 0 {
		code = int(ok)
	}
	if code > 600 || code < 0 {
		panic(fmt.Sprintf("Enum '%s' range must be greater than 0 and less than or equal to 600", string(enum.Desc.Name())))
	}
	var ew errorWrapper
	for _, v := range enum.Values {
		enumCode := code
		eCode := proto.GetExtension(v.Desc.Options(), errors.E_Code)
		if ok := eCode.(int32); ok != 0 {
			enumCode = int(ok)
		}
		// If the current enumeration does not contain 'errors.code'
		// or the code value exceeds the range, the current enum will be skipped
		if enumCode > 600 || enumCode < 0 {
			panic(fmt.Sprintf("Enum '%s' range must be greater than 0 and less than or equal to 600", string(v.Desc.Name())))
		}
		if enumCode == 0 {
			continue
		}

		comment := commentsString(v.Comments.Leading)
		if comment == "" {
			comment = commentsString(v.Comments.Trailing)
		}

		err := &errorInfo{
			Name:       string(enum.Desc.Name()),
			Value:      string(v.Desc.Name()),
			CamelValue: case2Camel(string(v.Desc.Name())),
			HTTPCode:   enumCode,
			Comment:    comment,
			HasComment: len(comment) > 0,
		}
		ew.Errors = append(ew.Errors, err)
	}
	if len(ew.Errors) == 0 {
		return true
	}
	g.P(ew.execute())

	return false
}

func case2Camel(name string) string {
	if !strings.Contains(name, "_") {
		if name == strings.ToUpper(name) {
			name = strings.ToLower(name)
		}
		return enCases.String(name)
	}
	strs := strings.Split(name, "_")
	words := make([]string, 0, len(strs))
	for _, w := range strs {
		hasLower := false
		for _, r := range w {
			if unicode.IsLower(r) {
				hasLower = true
				break
			}
		}
		if !hasLower {
			w = strings.ToLower(w)
		}
		w = enCases.String(w)
		words = append(words, w)
	}

	return strings.Join(words, "")
}

func commentsString(comments protogen.Comments) string {
	cmts := string(comments)
	if cmts == "" {
		return ""
	}
	var b []byte
	for _, line := range strings.Split(strings.TrimSuffix(string(cmts), "\n"), "\n") {
		b = append(b, line...)
	}
	return string(b)
}
