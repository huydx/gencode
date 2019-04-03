package schema

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

type visitor struct {
	schema  *Schema
	current *ast.TypeSpec
}

func (v *visitor) Visit(n ast.Node) ast.Visitor {
	switch n.(type) {
	case *ast.TypeSpec:
		v.current = n.(*ast.TypeSpec)
	case *ast.StructType:
		s := n.(*ast.StructType)
		st := v.constructStruct(s)
		v.schema.Structs = append(v.schema.Structs, st)
		return v
	}
	return v
}

func (v *visitor) constructStruct(node ast.Node) *Struct {
	fmt.Printf("current: %v\n", v.current)
	st := node.(*ast.StructType)
	sfields := make([]*Field, 0)
	for _, field := range st.Fields.List {
		f := &Field{}
		if len(field.Names) > 0 {
			fmt.Printf("add field %v\n", field.Names[0])
			f.Name = field.Names[0].Name
		}
		if field.Type != nil {
			fmt.Printf("add type %v\n", field.Type)
			f.Type = v.mapType(field.Type)
		}
		if field.Tag != nil {
			f.Tag = field.Tag.Value
		}
		sfields = append(sfields, f)
	}
	s := &Struct{
		Name:   v.current.Name.Name,
		Fields: sfields,
	}
	return s
}

func (v *visitor) mapType(t ast.Expr) Type {
	switch t.(type) {
	case *ast.Ident:
		fmt.Printf("ident: %v\n", t)
		id := t.(*ast.Ident)
		tname := id.Name
		if strings.Contains(tname, "int") {
			signed := false
			bits := int64(32)
			var err error
			if strings.HasPrefix(tname, "u") {
				signed = true
			}
			elems := strings.Split(tname, "int")
			if len(elems) == 2 {
				if elems[1] != "" {
					bits, err = strconv.ParseInt(elems[1], 10, 32)
					if err != nil {
						panic(err)
					}
				}
			} else {
				panic("not expect suffix")
			}
			return &IntType{
				Bits:   int(bits),
				Signed: signed,
			}
		}
		if strings.Contains(tname, "float") {
			elems := strings.Split(tname, "float")
			if len(elems) == 2 {
				fmt.Println(elems)
				bits, err := strconv.ParseInt(elems[1], 10, 32)
				if err != nil {
					panic(err)
				}
				return &FloatType{
					Bits: int(bits),
				}
			} else {
				panic("not expected suffix")
			}
		}
		if strings.Contains(tname, "string") {
			return &StringType{}
		}
	}
	return nil
}

func ParseSchema2(fname string) (*Schema, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fname, nil, 0)
	if err != nil {
		return nil, err
	}
	var v = &visitor{}
	v.schema = &Schema{}
	ast.Walk(v, f)
	fmt.Printf("schema %v\n", v.schema)
	return v.schema, nil
}
