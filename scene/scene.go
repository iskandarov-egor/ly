package scene

import (
	"ly/geo"
	"ly/img"
	"ly/spectra"
	"ly/debug"
	"ly/util/math32"
	"ly/sampling"
	"math"
	"math/rand"
	"fmt"
)

type ShapeHitPoint struct {
	Point geo.Vec3
	Normal geo.Vec3
	ShadingNormal geo.Vec3
	U float32
	V float32
	RayT float32
	Shading *Shading
	Shape Shape
	Dpdu geo.Vec3 // d(point)/d(textureU)
	Dpdv geo.Vec3 // d(point)/d(textureV)
	Dndu geo.Vec3 // d(normal)/d(textureU)
	Dndv geo.Vec3 // d(normal)/d(textureV)
}

type Shading struct {
	Material Material
	Glow spectra.Spectr
}

type Shape interface {
	RayIntersection(geo.Ray) (bool, *ShapeHitPoint)
	SamplePosition(sampler sampling.Sampler2D) (pos geo.Vec3, norm geo.Vec3, prob float32)
	SamplePdf(ray geo.Ray) float32
	BoundingBox() geo.Box
	Area() float32
}

type Scene struct {
	Shapes []Shape
	Lights []Light
	NonAreaLights []Light
	Accelerator Aggregate
	LightsPowerDistribution sampling.Distribution1D
}

func NewShading(mat Material, glow spectra.Spectr) *Shading {
	return &Shading{
		Material: mat,
		Glow: glow,
	}
}

func DefaultShading() *Shading {
	return NewShading(New1ColorMatteMaterial(0.3, 0.6, 1, 0, false), nil)
}

type Sphere struct {
	Shading *Shading
	Center geo.Vec3
	Radius float32
}

func MakeSphere(x, y, z float32, r float32) *Sphere {
	return &Sphere{
		Center: geo.Vec3{x, y, z},
		Radius: r,
	}
}

func (s *Sphere) SetShading(obj *Shading) {
	s.Shading = obj
}

func (s *Sphere) Add2Scene(scene *Scene) {
	if s.Shading == nil {
		s.Shading = DefaultShading()
	}
	scene.Shapes = append(scene.Shapes, s)
	if s.Shading.Glow != nil {
		light := NewAreaLight(s, s.Shading.Glow)
		scene.AddLight(light)
	}
}

func (s *Sphere) BoundingBox() (box geo.Box) {
	diag := geo.Vec3{s.Radius, s.Radius, s.Radius}
	return geo.Box{
		Min: s.Center.Sub(diag),
		Max: s.Center.Add(diag),
	}
}

func (s *Sphere) Area() float32 {
	return s.Radius * s.Radius * 4 * math.Pi
}

func (s *Sphere) SamplePosition(sampler sampling.Sampler2D) (ret geo.Vec3, norm geo.Vec3, prob float32) {
	panic("not impl")
}

func (s *Sphere) SamplePdf(ray geo.Ray) float32 {
	panic("not impl")
}

func (s *Sphere) RayIntersection(ray geo.Ray) (hit bool, hp *ShapeHitPoint) {
	d := ray.Direction
	o := ray.Origin.Sub(s.Center)
	a := d.X*d.X + d.Y*d.Y + d.Z*d.Z
	b := 2 * (d.X*o.X + d.Y*o.Y + d.Z*o.Z)
	c := o.X*o.X + o.Y*o.Y + o.Z*o.Z - (s.Radius)*(s.Radius)
	D := b*b - 4*a*c
	if D < 0 {
		return false, hp
	}
	Ds := float32(math.Sqrt(float64(D)))
	t1 := (-b + Ds) / (2 * a)
	t2 := (-b - Ds) / (2 * a)

	//  -----X------t1-----t2---->
	if t1 > t2 {
		t1, t2 = t2, t1
	}
	hp = &ShapeHitPoint{}
	hp.Shading = s.Shading
	hp.Shape = s
	hp.U = 0
	hp.V = 0 // TODO uv
	if t1 <= 0 {
		if t2 <= 0 {
			return false, hp
		} else {
			hp.Point = ray.At(t2)
			hp.Normal = hp.Point.Sub(s.Center).Normalized()
			hp.ShadingNormal = hp.Normal
			//hp.Point = hp.Point.Add(hp.Normal.Mul(0.001))
			hp.RayT = t2
			return true, hp
		}
	} else {
		hp.Point = ray.At(t1)
		hp.Normal = hp.Point.Sub(s.Center).Normalized()
		hp.ShadingNormal = hp.Normal
			//hp.Point = hp.Point.Add(hp.Normal.Mul(0.001))
		hp.RayT = t1
		return true, hp
	}
}

func RayIntersectShapes(shapes []Shape, ray geo.Ray) (ret *ShapeHitPoint) {
	for _, shape := range shapes {
		hit, hitPoint := shape.RayIntersection(ray)
		if hit && (ret == nil || hitPoint.RayT < ret.RayT) {
			ret = hitPoint
		}
	}
	return ret
}

func (s Scene) CastRay(ray geo.Ray) (ret *ShapeHitPoint) {
	if s.Accelerator != nil {
		return s.Accelerator.RayIntersection(ray)
	}
	ret = RayIntersectShapes(s.Shapes, ray)
	if ret != nil && ret.RayT != -1 {
		return ret
	} else {
		return nil
	}
}

// sample random light
// returns the light and the probability of sampling it
func (s Scene) SampleLight() (Light, float32) {
	if false {
		return s.Lights[rand.Intn(len(s.Lights))], 1/float32(len(s.Lights))
	} else {
		x, pdf := s.LightsPowerDistribution.Sample(rand.Float32())
		i := int(x * float32(len(s.Lights)))
		if i == len(s.Lights) {
			i--
		}
		return s.Lights[i], pdf/float32(len(s.Lights))
	}
}

func (s *Scene) Preprocess() {
	/* build light power disribution */
	{
		var lightPowers []float32
		for _, light := range s.Lights {
			pow := light.Power()
			lightPowers = append(lightPowers, pow)
		}
		s.LightsPowerDistribution, _ = sampling.NewDistribution1D(lightPowers)
	}
	/* init world bbox */
	bbox := geo.NewBox()
	for _, shape := range s.Shapes {
		bbox = bbox.Union(shape.BoundingBox())
	}
	/* init lights */
	{
		diag := bbox.Max.Sub(bbox.Min).Len()
		radius := diag / 2
		for _, light := range s.Lights {
			if l, ok := light.(*InfiniteAreaLight); ok {
				l.SetSceneRadius(radius)
			}
		}
	}
}

// the normal will point down
func MakePlane(x, y, z, w, h float32) *Mesh {
	var mesh Mesh
	w = w / 2
	h = h / 2
	// 1--3 ^ y
	// |  | |
	// 0--2 +--> x
	mesh.Vertices = []geo.Vec3{
		geo.Vec3{x - w, y - h, z},
		geo.Vec3{x - w, y + h, z},
		geo.Vec3{x + w, y - h, z},
		geo.Vec3{x + w, y + h, z},
	}
	mesh.Indices = []int{
		0, 1, 2,
		1, 3, 2,
	}
	mesh.U = []float32{0, 0, 1, 1}
	mesh.V = []float32{0, 1, 0, 1}
	return &mesh
}

func MakeCube(x, y, z, w float32, obj *Shading) *Mesh {
	var mesh Mesh
	mesh.Shading = obj
	w = w / 2
	/*
           2----3 6----7
	^ y    |    | |    |
	|      | z- | | z+ |
	+--> x 0----1 4----5 
	*/
	for i := 0; i < 8; i++ {
		var v geo.Vec3
		if i % 2 == 0 {
			v.X = x - w
		} else {
			v.X = x + w
		}
		if i % 4 < 2 {
			v.Y = y - w
		} else {
			v.Y = y + w
		}
		if i < 4 {
			v.Z = z - w
		} else {
			v.Z = z + w
		}
		mesh.Vertices = append(mesh.Vertices, v)
	}
	mesh.Vertices = append(mesh.Vertices, mesh.Vertices[6])
	mesh.Vertices = append(mesh.Vertices, mesh.Vertices[7])
	mesh.Vertices = append(mesh.Vertices, mesh.Vertices[2])
	mesh.Vertices = append(mesh.Vertices, mesh.Vertices[3])
	mesh.Vertices = append(mesh.Vertices, mesh.Vertices[0])
	mesh.Vertices = append(mesh.Vertices, mesh.Vertices[4])
	mesh.Indices = append(mesh.Indices, 5, 9, 8)
	mesh.Indices = append(mesh.Indices, 8, 4, 5)
	mesh.Indices = append(mesh.Indices, 11, 1, 0)
	mesh.Indices = append(mesh.Indices, 0, 10, 11)
	mesh.Indices = append(mesh.Indices, 2, 6, 7)
	mesh.Indices = append(mesh.Indices, 3, 2, 7)
	mesh.Indices = append(mesh.Indices, 3, 7, 5)
	mesh.Indices = append(mesh.Indices, 5, 1, 3)
	//if debug.Flag{
		mesh.Indices = append(mesh.Indices, 0, 5, 4)
		mesh.Indices = append(mesh.Indices, 5, 0, 1)
		mesh.Indices = append(mesh.Indices, 12, 13, 6)
		mesh.Indices = append(mesh.Indices, 6, 2, 12)
	//}
	/*
	8----9
	|    |
	|    |
	4----5----7----6----13
	|    |    |    |    |
	|    |    |    |    |
	0----1----3----2----12
	|    |
	|    |
	10---11
	*/
	mesh.U = []float32{0, 0.25, 0.75, 0.5, 0, 0.25, 0.75, 0.5, 0, 0.25, 0, 0.25, 1, 1}
	mesh.V = []float32{0.5, 0.5, 0.5, 0.5, 0.25, 0.25, 0.25, 0.25, 0, 0, 0.75, 0.75, 0.5, 0.25}
	return &mesh
}

func (s *Scene) AddLight(light Light) {
	s.Lights = append(s.Lights, light)
	if _, ok := light.(*AreaLight); !ok {
		s.NonAreaLights = append(s.NonAreaLights, light)
	}
}

type Triangle struct {
	Mesh *Mesh
	Idx  int
}

func (t *Triangle) BoundingBox() (box geo.Box) {
	box = geo.NewBox()
	box.Include(t.Mesh.Vertices[t.Mesh.Indices[t.Idx]])
	box.Include(t.Mesh.Vertices[t.Mesh.Indices[t.Idx + 1]])
	box.Include(t.Mesh.Vertices[t.Mesh.Indices[t.Idx + 2]])
	return
}

func (t *Triangle) RayIntersection(ray geo.Ray) (ok bool, hp *ShapeHitPoint) {
	m := t.Mesh
	i1, i2, i3 := m.Indices[t.Idx], m.Indices[t.Idx + 1], m.Indices[t.Idx + 2]
	p1 := m.Vertices[i1].Sub(ray.Origin)
	p2 := m.Vertices[i2].Sub(ray.Origin)
	p3 := m.Vertices[i3].Sub(ray.Origin)
	{
		/* permute coords so that ray.Direction.z has the greatest magnitude */
		x := math32.Abs(ray.Direction.X)
		y := math32.Abs(ray.Direction.Y)
		z := math32.Abs(ray.Direction.Z)
		if x > y {
			if x > z {
				p1.X, p1.Z = p1.Z, p1.X
				p2.X, p2.Z = p2.Z, p2.X
				p3.X, p3.Z = p3.Z, p3.X
				ray.Direction.X, ray.Direction.Z = ray.Direction.Z, ray.Direction.X
			}
		} else if y > z {
			p1.Y, p1.Z = p1.Z, p1.Y
			p2.Y, p2.Z = p2.Z, p2.Y
			p3.Y, p3.Z = p3.Z, p3.Y
			ray.Direction.Y, ray.Direction.Z = ray.Direction.Z, ray.Direction.Y
		}
	}

	sx := -ray.Direction.X/ray.Direction.Z
	sy := -ray.Direction.Y/ray.Direction.Z
	sz := 1/ray.Direction.Z
	{
		/* shear so that ray.Direction.z is (0, 0, 1) */
		p1.X += sx*p1.Z
		p1.Y += sy*p1.Z
		p2.X += sx*p2.Z
		p2.Y += sy*p2.Z
		p3.X += sx*p3.Z
		p3.Y += sy*p3.Z
		// postpone z shear till needed
	}
	/* compute the edge function coefficients */
	e1 := p2.X*p3.Y - p2.Y*p3.X
	e2 := p3.X*p1.Y - p3.Y*p1.X
	e3 := p1.X*p2.Y - p1.Y*p2.X
	if (e1 > 0 || e2 > 0 || e3 > 0) && (e1 < 0 || e2 < 0 || e3 < 0) {
		return false, hp
	}
	det := e1 + e2 + e3
	if det == 0 {
		return false, hp
	}
	op1 := m.Vertices[i1]
	op2 := m.Vertices[i2]
	op3 := m.Vertices[i3]
	{
		// z shear
		p1.Z  = sz*p1.Z
		p2.Z  = sz*p2.Z
		p3.Z  = sz*p3.Z
	}
	rayT := e1*p1.Z + e2*p2.Z + e3*p3.Z
	if (det > 0) != (rayT > 0) {
		return false, hp
	}
	invE := 1/det
	rayT *= invE
	b1 := e1 * invE
	b2 := e2 * invE
	b3 := e3 * invE
	hp = &ShapeHitPoint{}
	hp.RayT = rayT
	hp.Shading = m.Shading
	hp.Shape = t
	hp.Point = op1.Mul(b1).Add(op2.Mul(b2)).Add(op3.Mul(b3))
	hp.Normal = op1.Sub(op3).Cross(op2.Sub(op3)).Normalized()
	if len(m.Normals) == 0 {
		hp.ShadingNormal = hp.Normal
	} else {
		hp.ShadingNormal =
			m.Normals[i1].Mul(b1).Add(m.Normals[i2].Mul(b2)).Add(m.Normals[i3].Mul(b3)).Normalized()
	}
	hp.U = m.U[i1]*b1 + m.U[i2]*b2 + m.U[i3]*b3
	hp.V = m.V[i1]*b1 + m.V[i2]*b2 + m.V[i3]*b3

	/* bump mappin */
	/*
	if false {
		U13 := m.U[i3] - m.U[i1]
		U23 := m.U[i3] - m.U[i2]
		V13 := m.V[i3] - m.V[i1]
		V23 := m.V[i3] - m.V[i2]
		p13 := op3.Sub(op1)
		p23 := op3.Sub(op2)
		denom := 1/(U13*V23 - V13*U23)
		dsdu := p13.Mul(V23).Sub(p23.Mul(V13)).Mul(denom)
		dsdv := p13.Mul(U23).Sub(p23.Mul(U13)).Mul(denom)
		hp.Dndu = geo.Vec3{0, 0, 0} // TODO change this when normals come from a file!
		hp.Dndv = geo.Vec3{0, 0, 0} // TODO change this when normals come from a file!
		hp.Dpdu = dsdu
		hp.Dpdv = dsdv
	}
	*/

	return true, hp
}

func (t *Triangle) SamplePdf(ray geo.Ray) float32 {
	ok, hp := t.RayIntersection(ray)
	if !ok {
		return 0
	}
	dir := ray.Origin.Sub(hp.Point)
	dist2 := dir.LenSquared()
	dist := math32.Sqrt(dist2)
	if dist == 0 {
		return 0
	}
	cosDirNorm := dir.Scalar(hp.Normal) / dist
	if cosDirNorm < 0 {
		return 0
	}
	area := t.Area()
	prob := 1 / area
	if cosDirNorm == 0 {
		return 0
	}
	probAngle := prob*dist2/cosDirNorm
	return probAngle
}

func (t *Triangle) SamplePosition(sampler sampling.Sampler2D) (ret geo.Vec3, norm geo.Vec3, prob float32) {
	m := t.Mesh
	v1 := m.Vertices[m.Indices[t.Idx]]
	v2 := m.Vertices[m.Indices[t.Idx + 1]]
	v3 := m.Vertices[m.Indices[t.Idx + 2]]
	e1, e2 := sampler.Next()
	// u, v, s - barycentric coords of the sample
	u := 1 - math32.Sqrt(e1)
	v := e2*math32.Sqrt(e1)
	s := 1 - u - v
	ret = v1.Mul(u).Add(v2.Mul(v)).Add(v3.Mul(s))
	return
}

func (t *Triangle) Area() float32 {
	m := t.Mesh
	v1 := m.Vertices[m.Indices[t.Idx]]
	v2 := m.Vertices[m.Indices[t.Idx + 1]]
	v3 := m.Vertices[m.Indices[t.Idx + 2]]
	vec1 := v2.Sub(v1)
	vec2 := v3.Sub(v1)
	return 0.5*(vec1.Cross(vec2).Len())
}

type Actor interface {
	Add2Scene(scene *Scene, obj *Shading)
}

type Mesh struct {
	Shading    *Shading
	Vertices  []geo.Vec3
	Indices   []int
	Normals   []geo.Vec3
	U         []float32
	V         []float32
	area      float32
}

func (m *Mesh) Scale(scale geo.Vec3, center geo.Vec3) {
	x, y, z := scale.X, scale.Y, scale.Z
	for i, vertex := range m.Vertices {
		offset := vertex.Sub(center)
		offset.X *= x
		offset.Y *= y
		offset.Z *= z
		m.Vertices[i] = center.Add(offset)
	}
	for i := range m.Normals {
		m.Normals[i].Y *= y
		m.Normals[i].Z *= z
		m.Normals[i].X *= x
		m.Normals[i] = m.Normals[i].Normalized()
	}
}

func (m *Mesh) Translate(delta geo.Vec3) {
	for i := range m.Vertices {
		m.Vertices[i].X += delta.X
		m.Vertices[i].Y += delta.Y
		m.Vertices[i].Z += delta.Z
	}
}

func (m *Mesh) Rotate(axis geo.Vec3, angle float32) {
	cos := math32.Cos(angle)
	sin := math32.Sin(angle)
	x, y, z := axis.X, axis.Y, axis.Z
	for i, vertex := range m.Vertices {
		vx, vy, vz := vertex.X, vertex.Y, vertex.Z
		m.Vertices[i].X = vx*(cos + x*x*(1-cos)) + vy*(y*x*(1-cos) + z*sin) + vz*(z*x*(1-cos) - y*sin)
		m.Vertices[i].Y = vx*(x*y*(1-cos) - z*sin) + vy*(cos + y*y*(1-cos)) + vz*(z*y*(1-cos) + x*sin)
		m.Vertices[i].Z = vx*(x*z*(1-cos) + y*sin) + vy*(y*z*(1-cos) - x*sin) + vz*(cos + z*z*(1-cos))
	}
	for i, n := range m.Normals {
		vx, vy, vz := n.X, n.Y, n.Z
		m.Normals[i].X = vx*(cos + x*x*(1-cos)) + vy*(y*x*(1-cos) + z*sin) + vz*(z*x*(1-cos) - y*sin)
		m.Normals[i].Y = vx*(x*y*(1-cos) - z*sin) + vy*(cos + y*y*(1-cos)) + vz*(z*y*(1-cos) + x*sin)
		m.Normals[i].Z = vx*(x*z*(1-cos) + y*sin) + vy*(y*z*(1-cos) - x*sin) + vz*(cos + z*z*(1-cos))
	}
}

func (m *Mesh) SwapAxis(axis1, axis2 geo.Axis) {
	for i := range m.Vertices {
		*m.Vertices[i].AxisP(axis1), *m.Vertices[i].AxisP(axis2) = 
			m.Vertices[i].Axis(axis2), m.Vertices[i].Axis(axis1)
	}
}

func (m *Mesh) FlipNormals() {
	for i := 1; i < len(m.Indices); i += 3 {
		m.Indices[i], m.Indices[i + 1] = m.Indices[i + 1], m.Indices[i]
	}
	// TODO mirror m.Normals also
	if len(m.Normals) > 0 {
		println("flipping shading normals not implemented")
	}
}

func (m *Mesh) SetShading(obj *Shading) {
	m.Shading = obj
}

func (m *Mesh) Add2Scene(scene *Scene) {
	if len(m.Indices) % 3 != 0 {
		panic("bad mesh")
	}
	for i := 0; i < len(m.Indices); i += 3 {
		shape := &Triangle{
			Mesh: m,
			Idx: i,
		}
		scene.Shapes = append(scene.Shapes, shape)
		if m.Shading.Glow != nil {
			light := NewAreaLight(shape, m.Shading.Glow)
			scene.AddLight(light)
		}
	}
}

// 
// 
// 
// 
func barycentric(x, y, x0, y0, x1, y1, x2, y2 float32) (b0, b1, b2 float32) {
	invdet := 1/((x1 - x0)*(y2 - y0) - (y1 - y0)*(x2 - x0))
	b1 = invdet*((y2 - y0)*(x - x0) + (x0 - x2)*(y - y0))
	b2 = invdet*((y0 - y1)*(x - x0) + (x1 - x0)*(y - y0))
	b0 = 1 - b1 - b2
	return
}

func (m *Mesh) Uv2xyz(u, v float32) (ret, dpdu, dpdv geo.Vec3) {
	for idx := 0; idx < len(m.Indices); idx += 3 {
		i1, i2, i3 := m.Indices[idx], m.Indices[idx + 1], m.Indices[idx + 2]

		b0, b1, b2 :=
			barycentric(u, v, m.U[i1], m.V[i1], m.U[i2], m.V[i2], m.U[i3], m.V[i3])
		if b0 < 0 || b1 < 0 || b2 < 0 {
			continue
		}
		ret = m.Vertices[i1].Mul(b0).
			Add(m.Vertices[i2].Mul(b1)).
			Add(m.Vertices[i3].Mul(b2))
		op1 := m.Vertices[i1]
		op2 := m.Vertices[i2]
		op3 := m.Vertices[i3]
		U13 := m.U[i3] - m.U[i1]
		U23 := m.U[i3] - m.U[i2]
		V13 := m.V[i3] - m.V[i1]
		V23 := m.V[i3] - m.V[i2]
		p13 := op3.Sub(op1)
		p23 := op3.Sub(op2)
		denom := 1/(U13*V23 - V13*U23)
		dpdu = p13.Mul(V23).Sub(p23.Mul(V13)).Mul(denom)
		dpdv = p13.Mul(U23).Sub(p23.Mul(U13)).Mul(denom)
		return
	}
	fmt.Println(u, v, m.U[:1000], m.V[:1000])
	panic("achtung, uv2xyz fail")
	return
}

func init() {
	debug.Noop()
	img.Noop()
	if false { fmt.Println() }
	_ = rand.Intn
}
