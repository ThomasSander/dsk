// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday"
)

// Node represents a directory inside the design definitions tree.
type Node struct {
	path     string
	Title    string  `json:"title"`
	URL      string  `json:"url"`
	Parent   *Node   `json:"parent"`
	Children []*Node `json:"children"`
}

// Meta data as specified in a node configuration file.
type NodeMeta struct {
	// Optional, if missing will use the URL.
	Import string    `json:"import"`
	Demo   []PropSet `json:"demo"`
}

// A set of component properties, usually parsed from JSON.
type PropSet interface{}

// Constructs a new node using its path in the filesystem.
func NewNodeFromPath(path string, root string) *Node {
	var url string
	if path == root {
		url = "/"
	} else {
		url = strings.TrimSuffix(strings.TrimPrefix(path, root+"/"), "/")
	}
	return &Node{
		path:  path,
		URL:   url,
		Title: filepath.Base(path),
	}
}

// Recursively crawls the given root directory, constructing a flat list
// of nodes.
func NewNodeListFromPath(root string) ([]*Node, error) {
	var nodes []*Node

	err := filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			nodes = append(nodes, NewNodeFromPath(path, root))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build node tree in %s: %s", root, err)
	}

	for _, n := range nodes {
		for _, sn := range nodes {
			if filepath.Dir(sn.path) == n.path {
				n.Children = append(n.Children, sn)
			}
			// TODO: When adding parent, the conversion to JSON will create
			// recursions.
			// if filepath.Dir(n.path) == sn.path {
			// 	n.Parent = sn
			// }
		}
	}

	return nodes, nil
}

// Reads index.json file when present and returns values. When index.json
// is not present will simply return an empty Meta.
func (n Node) Meta() (NodeMeta, error) {
	var meta NodeMeta
	f := filepath.Join(n.path, "index.json")

	if _, err := os.Stat(f); os.IsNotExist(err) {
		return meta, nil
	}

	content, err := ioutil.ReadFile(f)
	if err != nil {
		return meta, err
	}
	if err := json.Unmarshal(content, &meta); err != nil {
		return meta, err
	}
	return meta, nil
}

// Result is passed as component import name to renderComponent()
// JavaScript glue function.
func (n Node) Import() (string, error) {
	m, err := n.Meta()
	if err != nil {
		return "", err
	}
	if m.Import != "" {
		return m.Import, nil
	}
	return n.URL, nil
}

func (n Node) CSS() (bytes.Buffer, error) {
	return n.bundledAssets("css")
}

func (n Node) JS() (bytes.Buffer, error) {
	return n.bundledAssets("js")
}

// Looks for i.e. CSS files in node directory and concatenates them.
// This way we don't need a naming convention for these assets.
func (n Node) bundledAssets(suffix string) (bytes.Buffer, error) {
	var b bytes.Buffer

	files, err := filepath.Glob(filepath.Join(n.path, "*."+suffix))
	if err != nil {
		return b, err
	}
	if len(files) == 0 {
		return b, fmt.Errorf("no .%s assets in path %s", suffix, n.path)
	}

	for _, f := range files {
		c, err := ioutil.ReadFile(f)
		if err != nil {
			return b, err
		}
		b.Write(c)
	}
	return b, nil
}

// Returns documentation parsed from markdown into HTML format.
func (n Node) Documentation() (template.HTML, error) {
	file := filepath.Join(n.path, "readme.md")

	if _, err := os.Stat(file); os.IsNotExist(err) {
		return "", nil
	}
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return template.HTML(blackfriday.Run(contents)), nil
}

// Returns a mapping of URLs to title strings for easily creating a
// breadcrumb navigation. The last element is the current active one.
// Does not include the very root element.
func (n Node) Crumbs() map[string]string {
	var crumbs map[string]string // maps url to title
	// TODO
	//	parts := strings.Split(n.url, "/")
	//
	//	for _, p := range parts {
	//		title := p
	//		url :=
	//	}
	return crumbs
}

func (n Node) GetDemos() ([]PropSet, error) {
	meta, err := n.Meta()
	if err != nil {
		return nil, err
	}
	return meta.Demo, nil
}

func (n Node) GetDemo(index int) (PropSet, error) {
	meta, err := n.Meta()
	if err != nil {
		return nil, err
	}
	return meta.Demo[index], nil
}