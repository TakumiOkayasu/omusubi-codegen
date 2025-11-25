package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TakumiOkayasu/omusubi-codegen/internal/model"
	"github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/cpp"
)

// Parser handles C++ code parsing using tree-sitter
type Parser struct {
	parser *sitter.Parser
}

// New creates a new C++ parser
func New() *Parser {
	parser := sitter.NewParser()
	lang := cpp.GetLanguage()
	if lang == nil {
		panic("failed to get C++ language from tree-sitter")
	}
	parser.SetLanguage(lang)
	return &Parser{
		parser: parser,
	}
}

// ParseFile parses a C++ header file and extracts class information
func (p *Parser) ParseFile(filePath string) (*model.FileInfo, error) {
	source, err := readFile(filePath)
	if err != nil {
		return nil, err
	}

	fileInfo, err := p.ParseSource(source)
	if err != nil {
		return nil, err
	}
	fileInfo.Path = filePath

	// Determine file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".h" {
		// Set extension for all classes in this file
		for i := range fileInfo.Classes {
			fileInfo.Classes[i].SourceFileExt = "h"
		}
	} else {
		// Default to hpp
		for i := range fileInfo.Classes {
			fileInfo.Classes[i].SourceFileExt = "hpp"
		}
	}

	return fileInfo, nil
}

// ParseSource parses C++ source code and extracts class information
func (p *Parser) ParseSource(source []byte) (*model.FileInfo, error) {
	ctx := context.Background()
	tree, err := p.parser.ParseCtx(ctx, nil, source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source: %w", err)
	}
	if tree == nil {
		return nil, fmt.Errorf("failed to parse source: tree is nil")
	}
	defer tree.Close()

	fileInfo := &model.FileInfo{
		Classes: []model.ClassInfo{},
	}

	rootNode := tree.RootNode()
	p.traverseNode(rootNode, source, fileInfo, "")

	return fileInfo, nil
}

// traverseNode recursively traverses the AST
func (p *Parser) traverseNode(node *sitter.Node, source []byte, fileInfo *model.FileInfo, currentNamespace string) {
	nodeType := node.Type()

	switch nodeType {
	case "namespace_definition":
		// Extract namespace name
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil {
			namespace := nameNode.Content(source)
			bodyNode := node.ChildByFieldName("body")
			if bodyNode != nil {
				p.traverseNode(bodyNode, source, fileInfo, namespace)
			}
		}
		return // Don't traverse children, we already handled the body

	case "class_specifier":
		classInfo := p.extractClassInfo(node, source, currentNamespace)
		if classInfo != nil && classInfo.IsAbstract {
			fileInfo.Classes = append(fileInfo.Classes, *classInfo)
			if fileInfo.Namespace == "" {
				fileInfo.Namespace = currentNamespace
			}
		}
		return // Don't traverse children, we already handled the class
	}

	// Traverse children
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		p.traverseNode(child, source, fileInfo, currentNamespace)
	}
}

// extractClassInfo extracts class information from a class_specifier node
func (p *Parser) extractClassInfo(node *sitter.Node, source []byte, namespace string) *model.ClassInfo {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	className := nameNode.Content(source)
	classInfo := &model.ClassInfo{
		Name:        className,
		Namespace:   namespace,
		BaseClasses: []string{},
		Methods:     []model.MethodInfo{},
		Fields:      []model.FieldInfo{},
		IsAbstract:  false,
	}

	// Extract base classes
	baseClassNode := node.ChildByFieldName("base_class_clause")
	if baseClassNode != nil {
		for i := 0; i < int(baseClassNode.ChildCount()); i++ {
			child := baseClassNode.Child(i)
			if child.Type() == "type_identifier" {
				classInfo.BaseClasses = append(classInfo.BaseClasses, child.Content(source))
			}
		}
	}

	// Extract body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		currentAccess := model.Private // Default for class
		p.extractClassMembers(bodyNode, source, classInfo, &currentAccess)
	}

	return classInfo
}

// extractClassMembers extracts methods and fields from class body
func (p *Parser) extractClassMembers(bodyNode *sitter.Node, source []byte, classInfo *model.ClassInfo, currentAccess *model.AccessLevel) {
	for i := 0; i < int(bodyNode.ChildCount()); i++ {
		child := bodyNode.Child(i)
		nodeType := child.Type()

		switch nodeType {
		case "access_specifier":
			accessText := child.Content(source)
			switch accessText {
			case "public:":
				*currentAccess = model.Public
			case "protected:":
				*currentAccess = model.Protected
			case "private:":
				*currentAccess = model.Private
			}

		case "function_definition", "declaration":
			method := p.extractMethodInfo(child, source, *currentAccess)
			if method != nil {
				classInfo.Methods = append(classInfo.Methods, *method)
				if method.IsPureVirtual {
					classInfo.IsAbstract = true
				}
			}

		case "field_declaration":
			field := p.extractFieldInfo(child, source, *currentAccess)
			if field != nil {
				classInfo.Fields = append(classInfo.Fields, *field)
			}
		}
	}
}

// extractMethodInfo extracts method information from a node
func (p *Parser) extractMethodInfo(node *sitter.Node, source []byte, access model.AccessLevel) *model.MethodInfo {
	declaratorNode := p.findNodeByType(node, "function_declarator")
	if declaratorNode == nil {
		return nil
	}

	// Extract method name
	nameNode := declaratorNode.ChildByFieldName("declarator")
	if nameNode == nil {
		return nil
	}

	methodName := ""
	if nameNode.Type() == "field_identifier" || nameNode.Type() == "identifier" {
		methodName = nameNode.Content(source)
	} else if nameNode.Type() == "destructor_name" {
		methodName = nameNode.Content(source)
	}

	if methodName == "" {
		return nil
	}

	method := &model.MethodInfo{
		Name:        methodName,
		Parameters:  []model.ParameterInfo{},
		AccessLevel: access,
	}

	// Extract return type
	typeNode := node.ChildByFieldName("type")
	if typeNode != nil {
		method.ReturnType = typeNode.Content(source)
	}

	// Extract parameters
	paramsNode := declaratorNode.ChildByFieldName("parameters")
	if paramsNode != nil {
		p.extractParameters(paramsNode, source, method)
	}

	// Check for virtual, pure virtual, const, static, override
	nodeContent := node.Content(source)
	method.IsVirtual = containsKeyword(nodeContent, "virtual")
	method.IsPureVirtual = containsKeyword(nodeContent, "= 0")
	method.IsConst = containsKeyword(nodeContent, "const")
	method.IsStatic = containsKeyword(nodeContent, "static")
	method.IsOverride = containsKeyword(nodeContent, "override")

	return method
}

// extractParameters extracts parameter information
func (p *Parser) extractParameters(paramsNode *sitter.Node, source []byte, method *model.MethodInfo) {
	for i := 0; i < int(paramsNode.ChildCount()); i++ {
		child := paramsNode.Child(i)
		if child.Type() == "parameter_declaration" {
			param := model.ParameterInfo{}

			typeNode := child.ChildByFieldName("type")
			if typeNode != nil {
				param.Type = typeNode.Content(source)
			}

			declaratorNode := child.ChildByFieldName("declarator")
			if declaratorNode != nil {
				param.Name = declaratorNode.Content(source)
			}

			defaultNode := child.ChildByFieldName("default_value")
			if defaultNode != nil {
				param.DefaultValue = defaultNode.Content(source)
			}

			method.Parameters = append(method.Parameters, param)
		}
	}
}

// extractFieldInfo extracts field information from a node
func (p *Parser) extractFieldInfo(node *sitter.Node, source []byte, access model.AccessLevel) *model.FieldInfo {
	typeNode := node.ChildByFieldName("type")
	if typeNode == nil {
		return nil
	}

	declaratorNode := node.ChildByFieldName("declarator")
	if declaratorNode == nil {
		return nil
	}

	fieldName := ""
	if nameNode := p.findNodeByType(declaratorNode, "field_identifier"); nameNode != nil {
		fieldName = nameNode.Content(source)
	}

	if fieldName == "" {
		return nil
	}

	field := &model.FieldInfo{
		Name:        fieldName,
		Type:        typeNode.Content(source),
		AccessLevel: access,
	}

	nodeContent := node.Content(source)
	field.IsStatic = containsKeyword(nodeContent, "static")
	field.IsConst = containsKeyword(nodeContent, "const")

	return field
}

// findNodeByType recursively finds a node by type
func (p *Parser) findNodeByType(node *sitter.Node, nodeType string) *sitter.Node {
	if node.Type() == nodeType {
		return node
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if result := p.findNodeByType(child, nodeType); result != nil {
			return result
		}
	}

	return nil
}

// containsKeyword checks if a string contains a keyword
func containsKeyword(content, keyword string) bool {
	return len(content) > 0 && len(keyword) > 0 && findSubstring(content, keyword)
}

// findSubstring is a simple substring search
func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ParseDirectory parses all header files in a directory and its subdirectories
func (p *Parser) ParseDirectory(dirPath string) ([]model.FileInfo, error) {
	var fileInfos []model.FileInfo
	var headerFiles []string

	// Find all .hpp and .h files
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".hpp" || ext == ".h" {
				headerFiles = append(headerFiles, path)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", dirPath, err)
	}

	// Parse each header file
	for _, headerFile := range headerFiles {
		fileInfo, err := p.ParseFile(headerFile)
		if err != nil {
			// Log error but continue parsing other files
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", headerFile, err)
			continue
		}

		// Only include files that contain abstract classes
		if len(fileInfo.Classes) > 0 {
			fileInfos = append(fileInfos, *fileInfo)
		}
	}

	return fileInfos, nil
}

// FindAbstractClass searches for a specific abstract class by name
func (p *Parser) FindAbstractClass(dirPath, className string) (*model.ClassInfo, error) {
	fileInfos, err := p.ParseDirectory(dirPath)
	if err != nil {
		return nil, err
	}

	for _, fileInfo := range fileInfos {
		for _, classInfo := range fileInfo.Classes {
			if classInfo.Name == className && classInfo.IsAbstract {
				return &classInfo, nil
			}
		}
	}

	return nil, fmt.Errorf("abstract class '%s' not found in directory %s", className, dirPath)
}

// readFile reads the content of a file
func readFile(filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return content, nil
}

// DetectWorkspace attempts to detect omusubi workspace structure
// Returns paths to core library and platform library if found
// If useLegacyName is true, searches for "pre-omusubi" instead of "omusubi"
func DetectWorkspace(startPath string, useLegacyName bool) (coreLibPath, platformLibPath string, err error) {
	// Determine core library directory name
	coreLibName := "omusubi"
	platformPrefix := "omusubi-"
	if useLegacyName {
		coreLibName = "pre-omusubi"
		platformPrefix = "pre-omusubi-"
	}

	// Try to find omusubi/pre-omusubi directory (core library)
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", "", err
	}

	// Search parent directories for workspace structure
	currentPath := absPath
	for {
		// Check if core library directory exists
		corePath := filepath.Join(currentPath, coreLibName)
		if dirExists(corePath) {
			// Check if include/omusubi exists (core library structure)
			if dirExists(filepath.Join(corePath, "include", "omusubi")) {
				coreLibPath = corePath

				// Look for platform library (e.g., omusubi-m5stack or pre-omusubi-m5stack)
				parentDir := currentPath
				entries, err := os.ReadDir(parentDir)
				if err == nil {
					for _, entry := range entries {
						if entry.IsDir() && strings.HasPrefix(entry.Name(), platformPrefix) && entry.Name() != "omusubi-codegen" && entry.Name() != "pre-omusubi-codegen" {
							platformPath := filepath.Join(parentDir, entry.Name())
							// Verify it has include directory
							if dirExists(filepath.Join(platformPath, "include")) {
								platformLibPath = platformPath
								break
							}
						}
					}
				}

				return coreLibPath, platformLibPath, nil
			}
		}

		// Move to parent directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached root directory
			break
		}
		currentPath = parentPath
	}

	if useLegacyName {
		return "", "", fmt.Errorf("workspace not found: could not locate pre-omusubi core library (alpha version)")
	}
	return "", "", fmt.Errorf("workspace not found: could not locate omusubi core library")
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}