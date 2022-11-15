// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This files are distributed under the General Public License v3.0.
// A copy of abovesaid license can be found in the LICENSE file.

// Package static_analysis implements the static analysis and metadata
// extraction functionalities for the Choreia project.
package static_analysis

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"

	"github.com/its-hmny/Choreia/go/metadata"
	log "github.com/sirupsen/logrus"
)

func init() {
	// Log as ASCII instead of the default JSON formatter.
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: "15:04:05"})
	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
	// Only log the warning severity or above.
	log.SetLevel(log.TraceLevel)
}

func ExtractFromPackage(pkg *ast.Package) (metadata.Package, error) {
	meta := metadata.Package{
		Name:      pkg.Name,
		Channels:  map[string]metadata.Channel{},
		Functions: map[string]metadata.Function{},
		InitFlow:  nil,
	}

	// TODO: Add imports expansion and recursive parsing

	for _, file := range pkg.Files {
		log.Trace("Found new file in package '%s'", pkg.Name)
		for _, fileDecl := range file.Decls {
			meta.Visit(fileDecl)
		}
	}

	return meta, nil
}

// Parses the given 'path' directory and extracts metadata.PackageMetadata
// from the resulting AST, if the parsing the fails the function bails out.
func Extract(path string) (map[string]metadata.Package, error) {
	// We want to ntercept all errors and fully resolve each Node
	flags := parser.DeclarationErrors | parser.SpuriousErrors
	// Parses the given directory/project and extracts a map of packages available.
	parsed, err := parser.ParseDir(token.NewFileSet(), path, nil, flags)
	if err != nil {
		log.Fatal(err)
	}

	extracted := make(map[string]metadata.Package)
	for _, pkg := range parsed {
		log.Trace("Found package: '%s'", pkg.Name)

		pkgMeta, err := ExtractFromPackage(pkg)
		extracted[pkgMeta.Name] = pkgMeta
		if err != nil {
			log.Fatal(err)
		}
	}

	return extracted, nil
}

// Extracts the 'metadata.PackageMetadata' and serializes them in JSON format,
// then writes a new file at 'outputPath'. Bails out at everytime it find an error.
func ExtractAndSave(inputPath, outputPath string) (int, error) {
	// Extracts metadata from the given program
	metadata, err := Extract(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	// Converts the received metadata to JSON format
	export, err := json.Marshal(metadata)
	if err != nil {
		log.Fatal(err)
	}

	// Creates or truncates the output .json files
	file, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Writes the extracted JSON content to said file
	bytes, err := file.Write(export)
	if err != nil {
		log.Fatal(err)
	}

	return bytes, nil
}
