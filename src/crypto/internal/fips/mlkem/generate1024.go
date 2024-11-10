// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

package main

import (
	"flag"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

var replacements = map[string]string{
	"k": "k1024",

	"CiphertextSize768":       "CiphertextSize1024",
	"EncapsulationKeySize768": "EncapsulationKeySize1024",

	"encryptionKey": "encryptionKey1024",
	"decryptionKey": "decryptionKey1024",

	"EncapsulationKey768":    "EncapsulationKey1024",
	"NewEncapsulationKey768": "NewEncapsulationKey1024",
	"parseEK":                "parseEK1024",

	"kemEncaps":  "kemEncaps1024",
	"pkeEncrypt": "pkeEncrypt1024",

	"DecapsulationKey768":    "DecapsulationKey1024",
	"NewDecapsulationKey768": "NewDecapsulationKey1024",
	"newKeyFromSeed":         "newKeyFromSeed1024",

	"kemDecaps":  "kemDecaps1024",
	"pkeDecrypt": "pkeDecrypt1024",

	"GenerateKey768": "GenerateKey1024",
	"generateKey":    "generateKey1024",

	"kemKeyGen": "kemKeyGen1024",
	"kemPCT":    "kemPCT1024",

	"encodingSize4":             "encodingSize5",
	"encodingSize10":            "encodingSize11",
	"ringCompressAndEncode4":    "ringCompressAndEncode5",
	"ringCompressAndEncode10":   "ringCompressAndEncode11",
	"ringDecodeAndDecompress4":  "ringDecodeAndDecompress5",
	"ringDecodeAndDecompress10": "ringDecodeAndDecompress11",
}

func main() {
	inputFile := flag.String("input", "", "")
	outputFile := flag.String("output", "", "")
	flag.Parse()

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, *inputFile, nil, parser.SkipObjectResolution|parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	cmap := ast.NewCommentMap(fset, f, f.Comments)

	// Drop header comments.
	cmap[ast.Node(f)] = nil

	// Remove top-level consts used across the main and generated files.
	var newDecls []ast.Decl
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok == token.CONST {
				continue // Skip const declarations
			}
			if d.Tok == token.IMPORT {
				cmap[decl] = nil // Drop pre-import comments.
			}
		}
		newDecls = append(newDecls, decl)
	}
	f.Decls = newDecls

	// Replace identifiers.
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			if replacement, ok := replacements[x.Name]; ok {
				x.Name = replacement
			}
		}
		return true
	})

	// Replace identifiers in comments.
	for _, c := range f.Comments {
		for _, l := range c.List {
			for k, v := range replacements {
				if k == "k" {
					continue
				}
				l.Text = strings.ReplaceAll(l.Text, k, v)
			}
		}
	}

	out, err := os.Create(*outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	out.WriteString("// Code generated by generate1024.go. DO NOT EDIT.\n\n")

	f.Comments = cmap.Filter(f).Comments()
	err = format.Node(out, fset, f)
	if err != nil {
		log.Fatal(err)
	}
}
