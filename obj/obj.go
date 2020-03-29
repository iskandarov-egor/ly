package obj

import (
	"ly/scene"
	"ly/geo"
	"ly/img"
	"fmt"
	"path"
	"os"
)

type ObjLoader struct {
	Scene *scene.Scene
}

func NewObjLoader(world *scene.Scene) *ObjLoader {
	return &ObjLoader{
		Scene: world,
	}
}

type ObjMesh struct {
	Mesh     *scene.Mesh
	Material scene.Material
	ObjectName string
}

type Obj struct {
	Meshes []ObjMesh
}

func (o *ObjLoader) LoadObj(filepath string) Obj {
	obj := ParseObj(filepath)
	objDir := path.Dir(filepath)
	mtlpath := filepath[:len(filepath) - len(path.Ext(filepath))] + ".mtl"
	var mtl *Mtl
	if _, err := os.Stat(mtlpath); err == nil {
		mtl, err = ParseMtl(ParseObj(mtlpath))
		if err != nil {
			panic(fmt.Errorf("load mtl %q: %v", mtlpath, err))
		}
		//fmt.Println(mtl)
	} else if !os.IsNotExist(err) {
		panic(fmt.Errorf("open %q: %v", mtlpath, err))
	}
	var mesh *scene.Mesh
	pushVertex := func(point FStatementPoint) {
		vStatement := obj.VStatements[point.V - 1]
		if point.Vt != 0 {
			vtStatement := obj.VtStatements[point.Vt - 1]
			u, v := vtStatement.U, vtStatement.V
			mesh.U = append(mesh.U, u)
			mesh.V = append(mesh.V, 1 - v)
		}
		if point.Vn != 0 {
			vn := *obj.VnStatements[point.Vn - 1]
			mesh.Normals = append(mesh.Normals, geo.Vec3{vn.X, vn.Y, vn.Z}.Normalized())
		}
		vertex := geo.Vec3{vStatement.X, vStatement.Y, vStatement.Z}
		mesh.Vertices = append(mesh.Vertices, vertex)
	}

	var usemtl string
	var objectName string

	ret := Obj{
		Meshes: []ObjMesh{},
	}

	commitMesh := func() {
		if len(mesh.Vertices) == 0 {
			return
		}
		//fmt.Println("commit mesh")
		if len(mesh.U) == 0 {
			//fmt.Println("with no vt")
			mesh.U = make([]float32, len(mesh.Vertices))
			mesh.V = make([]float32, len(mesh.Vertices))
		}
		if usemtl != "" {
			mat, ok := mtl.Materials[usemtl]
			if !ok {
				fmt.Println(mtl)
				panic(fmt.Sprintf("unknown material %q", usemtl))
			}
			if mat.MapKd != "" {
				im := img.LoadPng(path.Join(objDir, mat.MapKd))
				ret.Meshes = append(ret.Meshes, ObjMesh{
					Mesh: mesh,
					Material: scene.NewMatteMaterial(im, 0, false),
					ObjectName: objectName,
				})
			} else {
				ret.Meshes = append(ret.Meshes, ObjMesh{
					Mesh: mesh,
					Material: scene.New1ColorMatteMaterial(1, 1, 1, 0, false),
					ObjectName: objectName,
				})
			}
		} else {
			ret.Meshes = append(ret.Meshes, ObjMesh{
				Mesh: mesh,
				Material: scene.New1ColorMatteMaterial(1, 1, 1, 0, false),
				ObjectName: objectName,
			})
		}
	}

	for _, statementI := range obj.Statements {
		switch s := statementI.(type) {
			case *FStatement:
				if mesh == nil {
					mesh = &scene.Mesh{}
				}
				fanRoot := len(mesh.Vertices)
				pushVertex(s.Points[0])
				pushVertex(s.Points[1])
				for i := 2; i < len(s.Points); i++ {
					mesh.Indices = append(mesh.Indices, len(mesh.Vertices) - 1)
					pushVertex(s.Points[i])
					mesh.Indices = append(mesh.Indices, len(mesh.Vertices) - 1)
					mesh.Indices = append(mesh.Indices, fanRoot)
				}
			case *UsemtlStatement:
				if mtl == nil {
					panic(fmt.Sprintf("unexpected usemtl %v", s))
				}
				if mesh != nil {
					commitMesh()
				}
				mesh = &scene.Mesh{}
				usemtl = s.Name
				//fmt.Println("start mesh with mtl", s.Name, "and name", objectName)
			case *OStatement:
				if mesh != nil {
					commitMesh()
				}
				mesh = &scene.Mesh{}
				objectName = s.Name
				//fmt.Println("start mesh with mtl", usemtl, "and name", objectName)
		}
	}
	if mesh != nil {
		commitMesh()
	}
	return ret
	//fmt.Println(mesh)
	//o.Scene.AddMesh(&mesh, nil, scene.New1ColorMatteMaterial(1, 1, 1))
	//o.Scene.AddMesh(&mesh, nil, scene.NewDielectricMaterial())
}

func Noop() {}

func init() {
	_ = fmt.Print
}
