package main

type Document struct {
	Comment string           `json:"comment"`
	Route   map[string]Route `json:"route"`
}

type Route struct {
	Comment string `json:"comment"`
	GET     Node   `json:"get"`
	POST    Node   `json:"post"`
}

type Node struct {
	Prefix  string    `json:"prefix"`
	Method  string    `json:"method"`
	Comment string    `json:"comment"`
	Input   NodeField `json:"input"`
	Output  NodeField `json:"output"`
}

type NodeField struct {
	Comment   string         `json:"comment"`
	Arguments []NodeArgument `json:"arguments"`
}

type NodeArgument struct {
	Comment string `json:"comment"`
	Type    string `json:"type"`
}
