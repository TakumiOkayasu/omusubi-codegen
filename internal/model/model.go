package model

// ClassInfo represents a C++ class definition
type ClassInfo struct {
	Name          string
	Namespace     string
	BaseClasses   []string
	IsAbstract    bool
	Methods       []MethodInfo
	Fields        []FieldInfo
	Documentation string
}

// MethodInfo represents a C++ method
type MethodInfo struct {
	Name          string
	ReturnType    string
	Parameters    []ParameterInfo
	IsVirtual     bool
	IsPureVirtual bool
	IsConst       bool
	IsStatic      bool
	IsOverride    bool
	AccessLevel   AccessLevel
	Documentation string
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
	Classes   []ClassInfo
	Namespace string
}
