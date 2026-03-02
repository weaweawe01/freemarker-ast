package ast

// NodeType identifies template AST node kinds.
type NodeType string

const (
	NodeTypeRoot          NodeType = "Root"
	NodeTypeText          NodeType = "Text"
	NodeTypeInterpolation NodeType = "Interpolation"
	NodeTypeIf            NodeType = "If"
	NodeTypeIfBranch      NodeType = "IfBranch"
	NodeTypeAssignment    NodeType = "Assignment"
	NodeTypeAssignBlock   NodeType = "AssignBlock"
	NodeTypeMacro         NodeType = "Macro"
	NodeTypeFunction      NodeType = "Function"
	NodeTypeReturn        NodeType = "Return"
	NodeTypeList          NodeType = "List"
	NodeTypeItems         NodeType = "Items"
	NodeTypeSep           NodeType = "Sep"
	NodeTypeSwitch        NodeType = "Switch"
	NodeTypeSwitchBranch  NodeType = "SwitchBranch"
	NodeTypeOutputFormat  NodeType = "OutputFormat"
	NodeTypeAutoEsc       NodeType = "AutoEsc"
	NodeTypeNoAutoEsc     NodeType = "NoAutoEsc"
	NodeTypeAttempt       NodeType = "Attempt"
	NodeTypeUnifiedCall   NodeType = "UnifiedCall"
	NodeTypeComment       NodeType = "Comment"
	NodeTypeNested        NodeType = "Nested"
	NodeTypeArray         NodeType = "Array"
	NodeTypeMap           NodeType = "Map"
	NodeTypeIdentifier    NodeType = "Identifier"
	NodeTypeNumber        NodeType = "Number"
	NodeTypeString        NodeType = "String"
	NodeTypeBoolean       NodeType = "Boolean"
	NodeTypeBinary        NodeType = "Binary"
	NodeTypeUnary         NodeType = "Unary"
	NodeTypeBuiltin       NodeType = "Builtin"
	NodeTypeCall          NodeType = "Call"
	NodeTypeDot           NodeType = "Dot"
	NodeTypeDynamicKey    NodeType = "DynamicKey"
	NodeTypeDefaultTo     NodeType = "DefaultTo"
	NodeTypeExists        NodeType = "Exists"
	NodeTypeParenthetical NodeType = "Parenthetical"
)

// Node is a template AST node.
type Node interface {
	Type() NodeType
}

// Expr is an expression AST node.
type Expr interface {
	Node
	isExpr()
}

// Root is the template root.
type Root struct {
	Children []Node `json:"children"`
}

func (n *Root) Type() NodeType { return NodeTypeRoot }

// Text is a plain text chunk.
type Text struct {
	Value string `json:"value"`
}

func (n *Text) Type() NodeType { return NodeTypeText }

// Interpolation is ${...}, #{...}, or [=...].
type Interpolation struct {
	Opening string `json:"opening"`
	Expr    Expr   `json:"expr"`
}

func (n *Interpolation) Type() NodeType { return NodeTypeInterpolation }

// If is an if/elseif/else block.
type If struct {
	Branches []*IfBranch `json:"branches"`
	Else     []Node      `json:"else,omitempty"`
}

func (n *If) Type() NodeType { return NodeTypeIf }

// IfBranch is one if/elseif branch.
type IfBranch struct {
	Condition Expr   `json:"condition"`
	Children  []Node `json:"children"`
}

func (n *IfBranch) Type() NodeType { return NodeTypeIfBranch }

// Assignment is #assign/#global/#local with one or more assignment entries.
type Assignment struct {
	Scope     string            `json:"scope"`
	Items     []*AssignmentItem `json:"items"`
	Namespace Expr              `json:"namespace,omitempty"`
}

func (n *Assignment) Type() NodeType { return NodeTypeAssignment }

// AssignmentItem is one target/op/value tuple.
type AssignmentItem struct {
	Target string `json:"target"`
	Op     string `json:"op"`
	Value  Expr   `json:"value,omitempty"`
}

// AssignBlock is capture-style assignment (<#assign x>...</#assign>).
type AssignBlock struct {
	Scope     string `json:"scope"`
	Target    string `json:"target"`
	Namespace Expr   `json:"namespace,omitempty"`
	Children  []Node `json:"children"`
}

func (n *AssignBlock) Type() NodeType { return NodeTypeAssignBlock }

// Macro is a minimal macro block.
type Macro struct {
	Name       string      `json:"name"`
	Params     []*ParamDef `json:"params,omitempty"`
	CatchAll   string      `json:"catch_all,omitempty"`
	IsFunction bool        `json:"is_function,omitempty"`
	Children   []Node      `json:"children"`
}

func (n *Macro) Type() NodeType { return NodeTypeMacro }

// Function is #function ... </#function>.
type Function struct {
	Name     string      `json:"name"`
	Params   []*ParamDef `json:"params,omitempty"`
	CatchAll string      `json:"catch_all,omitempty"`
	Children []Node      `json:"children"`
}

func (n *Function) Type() NodeType { return NodeTypeFunction }

// Return is #return expression or #return.
type Return struct {
	Value Expr `json:"value,omitempty"`
}

func (n *Return) Type() NodeType { return NodeTypeReturn }

// List is #list ... </#list>.
type List struct {
	Source   Expr   `json:"source"`
	LoopVar  string `json:"loop_var,omitempty"`
	Children []Node `json:"children,omitempty"`
	Else     []Node `json:"else,omitempty"`
}

func (n *List) Type() NodeType { return NodeTypeList }

// Items is #items ... </#items>.
type Items struct {
	LoopVar  string `json:"loop_var"`
	Children []Node `json:"children,omitempty"`
}

func (n *Items) Type() NodeType { return NodeTypeItems }

// Sep is #sep body.
type Sep struct {
	Children []Node `json:"children,omitempty"`
}

func (n *Sep) Type() NodeType { return NodeTypeSep }

// Switch is #switch ... (#case/#on/#default) ... </#switch>.
type Switch struct {
	Value    Expr            `json:"value"`
	Branches []*SwitchBranch `json:"branches,omitempty"`
	Default  []Node          `json:"default,omitempty"`
}

func (n *Switch) Type() NodeType { return NodeTypeSwitch }

// SwitchBranch is one #case or #on branch.
type SwitchBranch struct {
	Kind       string `json:"kind"` // "case" or "on"
	Conditions []Expr `json:"conditions,omitempty"`
	Children   []Node `json:"children,omitempty"`
}

func (n *SwitchBranch) Type() NodeType { return NodeTypeSwitchBranch }

// OutputFormat is #outputFormat value ... </#outputFormat>.
type OutputFormat struct {
	Value    Expr   `json:"value"`
	Children []Node `json:"children,omitempty"`
}

func (n *OutputFormat) Type() NodeType { return NodeTypeOutputFormat }

// AutoEsc is #autoEsc ... </#autoEsc>.
type AutoEsc struct {
	Children []Node `json:"children,omitempty"`
}

func (n *AutoEsc) Type() NodeType { return NodeTypeAutoEsc }

// NoAutoEsc is #noAutoEsc ... </#noAutoEsc>.
type NoAutoEsc struct {
	Children []Node `json:"children,omitempty"`
}

func (n *NoAutoEsc) Type() NodeType { return NodeTypeNoAutoEsc }

// Attempt is #attempt ... #recover ... </#attempt>.
type Attempt struct {
	Attempt []Node `json:"attempt,omitempty"`
	Recover []Node `json:"recover,omitempty"`
}

func (n *Attempt) Type() NodeType { return NodeTypeAttempt }

// Comment is a template comment block.
type Comment struct {
	Content string `json:"content"`
}

func (n *Comment) Type() NodeType { return NodeTypeComment }

// Nested is #nested directive with optional passed values.
type Nested struct {
	Values []Expr `json:"values,omitempty"`
}

func (n *Nested) Type() NodeType { return NodeTypeNested }

// UnifiedCall is user-defined directive call, like <@foo x=1; b1, b2>...</@foo>.
type UnifiedCall struct {
	Callee     Expr        `json:"callee"`
	Positional []Expr      `json:"positional,omitempty"`
	Named      []*NamedArg `json:"named,omitempty"`
	LoopVars   []string    `json:"loop_vars,omitempty"`
	Children   []Node      `json:"children,omitempty"`
}

func (n *UnifiedCall) Type() NodeType { return NodeTypeUnifiedCall }

// NamedArg is one named argument in a unified call header.
type NamedArg struct {
	Name  string `json:"name"`
	Value Expr   `json:"value"`
}

// ParamDef is one macro/function parameter definition.
type ParamDef struct {
	Name    string `json:"name"`
	Default Expr   `json:"default,omitempty"`
}

// Array is [...] literal.
type Array struct {
	Items []Expr `json:"items,omitempty"`
}

func (n *Array) Type() NodeType { return NodeTypeArray }
func (n *Array) isExpr()        {}

// Map is {...} literal.
type Map struct {
	Items []*MapEntry `json:"items,omitempty"`
}

func (n *Map) Type() NodeType { return NodeTypeMap }
func (n *Map) isExpr()        {}

// MapEntry is one key:value pair.
type MapEntry struct {
	Key   Expr `json:"key"`
	Value Expr `json:"value"`
}

// Identifier expression.
type Identifier struct {
	Name string `json:"name"`
}

func (n *Identifier) Type() NodeType { return NodeTypeIdentifier }
func (n *Identifier) isExpr()        {}

// Number expression.
type Number struct {
	Literal string `json:"literal"`
}

func (n *Number) Type() NodeType { return NodeTypeNumber }
func (n *Number) isExpr()        {}

// String literal expression.
type String struct {
	Literal string `json:"literal"`
}

func (n *String) Type() NodeType { return NodeTypeString }
func (n *String) isExpr()        {}

// Boolean literal expression.
type Boolean struct {
	Value bool `json:"value"`
}

func (n *Boolean) Type() NodeType { return NodeTypeBoolean }
func (n *Boolean) isExpr()        {}

// Binary expression.
type Binary struct {
	Op    string `json:"op"`
	Left  Expr   `json:"left"`
	Right Expr   `json:"right"`
}

func (n *Binary) Type() NodeType { return NodeTypeBinary }
func (n *Binary) isExpr()        {}

// Unary expression.
type Unary struct {
	Op   string `json:"op"`
	Expr Expr   `json:"expr"`
}

func (n *Unary) Type() NodeType { return NodeTypeUnary }
func (n *Unary) isExpr()        {}

// Builtin expression (exp?name(args...)).
type Builtin struct {
	Target Expr   `json:"target"`
	Name   string `json:"name"`
	Args   []Expr `json:"args,omitempty"`
}

func (n *Builtin) Type() NodeType { return NodeTypeBuiltin }
func (n *Builtin) isExpr()        {}

// Call expression (target(args...)).
type Call struct {
	Target Expr   `json:"target"`
	Args   []Expr `json:"args,omitempty"`
}

func (n *Call) Type() NodeType { return NodeTypeCall }
func (n *Call) isExpr()        {}

// Dot expression (target.name).
type Dot struct {
	Target Expr   `json:"target"`
	Name   string `json:"name"`
}

func (n *Dot) Type() NodeType { return NodeTypeDot }
func (n *Dot) isExpr()        {}

// DynamicKey expression (target[key]).
type DynamicKey struct {
	Target Expr `json:"target"`
	Key    Expr `json:"key"`
}

func (n *DynamicKey) Type() NodeType { return NodeTypeDynamicKey }
func (n *DynamicKey) isExpr()        {}

// DefaultTo expression (target!rhs or target!).
type DefaultTo struct {
	Target Expr `json:"target"`
	RHS    Expr `json:"rhs,omitempty"`
}

func (n *DefaultTo) Type() NodeType { return NodeTypeDefaultTo }
func (n *DefaultTo) isExpr()        {}

// Exists expression (target??).
type Exists struct {
	Target Expr `json:"target"`
}

func (n *Exists) Type() NodeType { return NodeTypeExists }
func (n *Exists) isExpr()        {}

// Parenthetical wraps an expression that was enclosed in parentheses.
type Parenthetical struct {
	Expr Expr `json:"expr"`
}

func (n *Parenthetical) Type() NodeType { return NodeTypeParenthetical }
func (n *Parenthetical) isExpr()        {}
