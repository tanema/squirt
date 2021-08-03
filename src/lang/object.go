package lang

type NodeKind string

const (
	Invalid    NodeKind = "<invalid>"
	Assignment NodeKind = "assign"
	AttrDef    NodeKind = "attr"
	Binary     NodeKind = "binary"
	Bool       NodeKind = "bool"
	Break      NodeKind = "break"
	ClassDef   NodeKind = "classdef"
	Cleanup    NodeKind = "cleanup"
	Do         NodeKind = "do"
	ForIn      NodeKind = "forin"
	ForNum     NodeKind = "fornum"
	FuncCall   NodeKind = "funccall"
	FuncDef    NodeKind = "funcdef"
	Identifier NodeKind = "identifier"
	If         NodeKind = "if"
	IfClause   NodeKind = "ifclause"
	Index      NodeKind = "index"
	Member     NodeKind = "member"
	Next       NodeKind = "next"
	Nil        NodeKind = "nil"
	Number     NodeKind = "number"
	Range      NodeKind = "range"
	Return     NodeKind = "return"
	Root       NodeKind = "root"
	Spread     NodeKind = "spread"
	String     NodeKind = "string"
	Table      NodeKind = "table"
	TableKey   NodeKind = "tablekey"
	TableValue NodeKind = "tablevalue"
	Ternary    NodeKind = "ternary"
	Unary      NodeKind = "unary"
	While      NodeKind = "while"
)

var invalid = Object{Kind: Invalid}

type (
	Object struct {
		Kind        NodeKind `json:"kind"`
		Name        string   `json:"name,omitempty"`
		Parent      string   `json:"parent,omitempty"`
		NumberValue float64  `json:"number,omitempty"`
		StringValue string   `json:"string,omitempty"`
		BoolValue   bool     `json:"bool,omitempty"`
		Cond        *Object  `json:"condition,omitempty"`
		Step        *Object  `json:"step,omitempty"`
		Key         *Object  `json:"key,omitempty"`
		Value       *Object  `json:"value,omitempty"`
		Vars        []Object `json:"variables,omitempty"`
		Vals        []Object `json:"values,omitempty"`
		Block       []Object `json:"block,omitempty"`
		Catches     []Object `json:"catches,omitempty"`
		Private     bool     `json:"private,omitempty"`
		Static      bool     `json:"static,omitempty"`
		Pos         [4]int   `json:"position"`
	}
)
