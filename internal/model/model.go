package model

// ClassInfo represents a C++ class definition
type ClassInfo struct {
	Name          string
	Namespace     string
	Documentation string
	SourceFileExt string // "h" or "hpp"
	BaseClasses   []string
	Methods       []MethodInfo
	Fields        []FieldInfo
	IsAbstract    bool
}

// MethodInfo represents a C++ method
type MethodInfo struct {
	Name          string
	ReturnType    string
	Documentation string
	Parameters    []ParameterInfo
	AccessLevel   AccessLevel
	IsVirtual     bool
	IsPureVirtual bool
	IsConst       bool
	IsStatic      bool
	IsOverride    bool
}

// ParameterInfo represents a method parameter
type ParameterInfo struct {
	Name         string
	Type         string
	DefaultValue string
}

// FieldInfo represents a class member variable
type FieldInfo struct {
	Name        string
	Type        string
	AccessLevel AccessLevel
	IsStatic    bool
	IsConst     bool
}

// AccessLevel represents C++ access modifiers
type AccessLevel int

const (
	Public AccessLevel = iota
	Protected
	Private
)

func (a AccessLevel) String() string {
	switch a {
	case Public:
		return "public"
	case Protected:
		return "protected"
	case Private:
		return "private"
	default:
		return "unknown"
	}
}

// FileInfo represents parsed C++ file information
type FileInfo struct {
	Path      string
	Namespace string
	Classes   []ClassInfo
}
