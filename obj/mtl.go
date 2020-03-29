package obj

import (
	"fmt"
)

type Material struct {
	MapKd string
	Name  string
}

type Mtl struct {
	Materials map[string]Material
}

func ParseMtl(mtl *ObjFile) (*Mtl, error) {
	ret := Mtl{Materials: make(map[string]Material)}
	var curMaterial *Material
	for _, statementI := range mtl.Statements {
		switch s := statementI.(type) {
			case *NewmtlStatement:
				if curMaterial != nil {
					ret.Materials[curMaterial.Name] = *curMaterial
				}
				curMaterial = &Material{Name: s.Name}
			case *MapKdStatement:
				if curMaterial == nil {
					return nil, fmt.Errorf("unexpected map_Kd %q", s.Path)
				}
				curMaterial.MapKd = s.Path
		}
	}
	if curMaterial != nil {
		ret.Materials[curMaterial.Name] = *curMaterial
	}
	return &ret, nil
}
