package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"fmt"
	"math"
	"sort"
	"reflect"
	"ly/scene"
	"ly/img"
	"ly/geo"
	"ly/cameras"
	"ly/obj"
	"ly/spectra"
	"ly/tracers"
)

type Typed struct {
	Type string `yaml:"type"`
}

func DecodeType(node *yaml.Node) (string, error) {
	var typed Typed
	err := node.Decode(&typed)
	if err != nil {
		return "", fmt.Errorf("decode type: %v", err)
	}
	if typed.Type == "" {
		return "", fmt.Errorf("type missing")
	}
	return typed.Type, nil
}

type MaterialConfig struct {
	Typed
}

type MatteMaterialConfig struct {
	MaterialConfig
	Texture       *string       `yaml:"texture"`
	Color         *VectorConfig `yaml:"color"`
	Roughness     float32       `yaml:"roughness"`
	IsTransparent bool          `yaml:"is_transparent"`
}

type LayerMaterialConfig struct {
	MatteMaterialConfig `yaml:",inline"`
	Eta    *float32     `yaml:"refractive_index"`
	Base   yaml.Node    `yaml:"base"`
}

type WeighedSumMaterialConfig struct {
	MaterialConfig
	Materials   []yaml.Node `yaml:"materials"`
	Weights     []float32   `yaml:"weights"`
}

type BlendMapMaterialConfig struct {
	MaterialConfig
	Black   yaml.Node  `yaml:"black"`
	White   yaml.Node  `yaml:"white"`
	Map     string     `yaml:"map"`
}

type MetalMaterialConfig struct {
	MaterialConfig
	K         *VectorConfig `yaml:"absorption_coefficient"`
	Roughness float32       `yaml:"roughness"`
	Color     *VectorConfig `yaml:"color"`
}

type DielectricMaterialConfig struct {
	MaterialConfig `yaml:",inline"`
	Color             *VectorConfig `yaml:"color"`
	ReflectionColor   *VectorConfig `yaml:"reflection_color"`
	Eta               *float32      `yaml:"refractive_index"`
	Roughness         *float32      `yaml:"roughness"`
}

type PlasticMaterialConfig struct {
	MaterialConfig
	Ks  *VectorConfig `yaml:"ks"`
	Kd  *VectorConfig `yaml:"kd"`
	Eta *float32 `yaml:"refractive_index"`
	Roughness float32 `yaml:"roughness"`
}

type ObjectConfig struct {
	Typed
	Material string `yaml:"material"`
	Glow *VectorConfig `yaml:"glow"`
}

type BoxObjectConfig struct {
	ObjectConfig `yaml:",inline"`
	Center *VectorConfig `yaml:"center"`
	Width  *float32 `yaml:"width"`
	Transformation *TransformationConfig `yaml:"transformation"`
}

type SphereObjectConfig struct {
	ObjectConfig `yaml:",inline"`
	Position *VectorConfig `yaml:"position"`
	Radius  *float32 `yaml:"radius"`
}

type PlaneObjectConfig struct {
	ObjectConfig `yaml:",inline"`
	Position    *VectorConfig    `yaml:"position"`
	Size        *TwoFloatsConfig `yaml:"size"`
    Orientation *string `yaml:"orientation"`
}

type RotationConfig struct {
	Axis VectorConfig `yaml:"axis"`
	Angle *float32 `yaml:"angle"`
}

type ObjObjectConfig struct {
	ObjectConfig `yaml:",inline"`
	Path *string `yaml:"path"`
	Transformation *TransformationConfig `yaml:"transformation"`
	OverrideMaterials map[string]string  `yaml:"override_materials"`
	OverrideGlow      map[string]*VectorConfig  `yaml:"override_glow"`
}

type CameraConfig struct {
	Typed
}

type PerspectiveCameraConfig struct {
	CameraConfig
	Zoom     *float32      `yaml:"zoom"`
	Fov      *float32      `yaml:"fov"`
	Position *VectorConfig `yaml:"position"`
	Target   *VectorConfig `yaml:"target"`
}

type LightConfig struct {
	Typed
}

type DirectionalLightConfig struct {
	LightConfig
	Direction *VectorConfig `yaml:"direction"`
	Color     *VectorConfig `yaml:"color"`
}

type InfiniteAreaLightConfig struct {
	LightConfig
	Scale     *float32 `yaml:"scale"`
	Direction *float32 `yaml:"direction"`
	Texture   *string `yaml:"texture"`
}

type TransformationConfig []map[string]yaml.Node

type VectorConfig struct {
	geo.Vec3
}

func (v *VectorConfig) UnmarshalYAML(node *yaml.Node) error {
	var list []float32
	err := node.Decode(&list)
	if err != nil {
		return err
	}
	if len(list) != 3 {
		return fmt.Errorf("expected a list of 3 floats, got %q", v)
	}
	v.X = list[0]
	v.Y = list[1]
	v.Z = list[2]
	return nil
}

func (v *VectorConfig) ToSpectr() spectra.Spectr {
	return spectra.NewRGBSpectr(v.X, v.Y, v.Z)
}

type TwoFloatsConfig [2]float32

func (v *TwoFloatsConfig) UnmarshalYAML(node *yaml.Node) error {
	var list []float32
	err := node.Decode(&list)
	if err != nil {
		return err
	}
	if len(list) != 2 {
		return fmt.Errorf("expected a list of 2 floats, got %q", v)
	}
	v[0] = list[0]
	v[1] = list[1]
	return nil
}

type SceneConfigYaml struct {
	Options `yaml:",inline"`
	Materials map[string]yaml.Node `yaml:"materials"`
	Objects   map[string]yaml.Node `yaml:"objects"`
	Cameras   map[string]yaml.Node `yaml:"cameras"`
	Lights    map[string]yaml.Node `yaml:"lights"`
	ActiveCamera string `yaml:"active_camera"`
	Accelerator  string `yaml:"accelerator"`
	Profile   string `yaml:"profile"`
	Profiles  map[string]ProfileConfig `yaml:"profiles"`
}

type TracerConfig struct {
	Typed `yaml:",inline"`
}

type PathTracerConfig struct {
	TracerConfig `yaml:",inline"`
	MinDepth int `yaml:"min_depth"`
	TerminationProb float32 `yaml:"termination_prob"`
}

type FTLTracerConfig struct {
	TracerConfig          `yaml:",inline"`
	MinDepth int          `yaml:"min_depth"`
	TerminationProb float32 `yaml:"termination_prob"`
	NFrames int           `yaml:"n_frames"`
	Fps float32           `yaml:"fps"`
	TimeOffset float32    `yaml:"time_offset"`
	LightDuration float32 `yaml:"light_duration"`
	SkipFirstSegment bool `yaml:"skip_first_segment"`
	OutfilePattern   string `yaml:"outfile_pattern"`
}

type ProfileConfig struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
	PixelSamples int `yaml:"pixel_samples"`
	SaveInterval int `yaml:"save_interval"`
	Tracer yaml.Node `yaml:"tracer"`
}

type Options struct {
	Profile ProfileConfig `yaml:"topsecret"` // not shadowing "profile" from SceneConfig
	Goroutines int  `yaml:"goroutines"`
	Outfile string  `yaml:"outfile"`
	Region  [4]int  `yaml:"region"`
}

type SceneConfig struct {
	Options *Options
	Camera cameras.Camera
	Tracer tracers.Tracer
	FTLTracer *tracers.FTLTracer
}

type MaterialMap struct {
	Map map[string]scene.Material
	Default scene.Material
}

func (m *MaterialMap) Get(key string) (scene.Material, error) {
	if key == "mirror" {
		return scene.NewMirrorMaterial(), nil
	}
	if key == "gold" {
		return scene.NewFourierMaterial("../pbrt-v3-scenes/bsdfs/roughgold_alpha_0.2.bsdf"), nil
	}
	if key == "l" {
		return scene.NewLayeredMaterial(m.Default.(*scene.MatteMaterial), 1.5), nil
	}
	mat, ok := m.Map[key]
	if ok {
		return mat, nil
	} else {
		return nil, fmt.Errorf("no such material: %q", key)
	}
}

func (m *MaterialMap) GetWithDefault(key string) scene.Material {
	if m.Default == nil {
		panic("aaa")
	}
	mat, err := m.Get(key)
	if err != nil {
		return m.Default
	}
	return mat
}

func ApplyTransformation(conf TransformationConfig, mesh *scene.Mesh) error {
	for _, transform := range conf {
		for k, tnode := range transform {
			switch k {
				case "translate":
					var t VectorConfig
					err := tnode.Decode(&t)
					if err != nil {
						return fmt.Errorf("bad %q transformation", k)
					}
					mesh.Translate(t.Vec3)
				case "scale":
					var t VectorConfig
					err := tnode.Decode(&t)
					if err != nil {
						return fmt.Errorf("bad %q transformation", k)
					}
					mesh.Scale(t.Vec3, geo.Vec3{0, 0, 0})
				case "flip":
					mesh.FlipNormals()
				case "swap":
					mesh.SwapAxis(0, 1)
				case "rotate":
					var t RotationConfig
					err := tnode.Decode(&t)
					if err != nil {
						return fmt.Errorf("bad %q transformation: %s", k, err)
					}
					if t.Angle == nil {
						return fmt.Errorf("rotate transformation: angle required")
					}
					zero := VectorConfig{geo.Vec3{0, 0, 0}}
					if t.Axis == zero {
						t.Axis.Z = 1
					}
					mesh.Rotate(t.Axis.Vec3, (*t.Angle)*math.Pi/180)
			}
		}
	}
	return nil
}

func LoadObj(node *yaml.Node, world *scene.Scene, matMap MaterialMap) error {
	var cfg ObjObjectConfig
	err := node.Decode(&cfg)
	if err != nil {
		return err
	}
	if cfg.Path == nil {
		return fmt.Errorf("path required")
	}
	loader := obj.NewObjLoader(world)
	objFile := loader.LoadObj(*cfg.Path)
	if cfg.Transformation != nil {
		for _, mesh := range objFile.Meshes {
			err := ApplyTransformation(*cfg.Transformation, mesh.Mesh)
			if err != nil {
				return err
			}
		}
	}
	defaultShading := scene.Shading{
		Material: matMap.GetWithDefault(cfg.Material),
	}
	if cfg.Glow != nil {
		defaultShading.Glow = cfg.Glow.ToSpectr()
	}
	for _, mesh := range objFile.Meshes {
		shading := defaultShading
		if cfg.OverrideMaterials != nil && cfg.OverrideMaterials[mesh.ObjectName] != "" {
			material, err := matMap.Get(cfg.OverrideMaterials[mesh.ObjectName])
			if err != nil {
				return err
			}
			shading.Material = material
		}
		if cfg.OverrideGlow != nil && cfg.OverrideGlow[mesh.ObjectName] != nil {
			shading.Glow = cfg.OverrideGlow[mesh.ObjectName].ToSpectr()
		}
		mesh.Mesh.SetShading(&shading)
		mesh.Mesh.Add2Scene(world)
	}
	return nil
}

func LoadBox(node *yaml.Node, world *scene.Scene, matMap MaterialMap) error {
	var cfg BoxObjectConfig
	err := node.Decode(&cfg)
	if err != nil {
		return err
	}
	if cfg.Center == nil || cfg.Width == nil {
		return fmt.Errorf("center and width are required")
	}
	material := matMap.GetWithDefault(cfg.Material)
	obj := &scene.Shading{
		Material: material,
	}
	if cfg.Glow != nil {
		obj.Glow = spectra.NewRGBSpectr(cfg.Glow.X, cfg.Glow.Y, cfg.Glow.Z)
	}
	box := scene.MakeCube(cfg.Center.X, cfg.Center.Y, cfg.Center.Z, *cfg.Width, obj)
	if cfg.Transformation != nil {
		err := ApplyTransformation(*cfg.Transformation, box)
		if err != nil {
			return err
		}
	}
	box.SetShading(obj)
	box.Add2Scene(world)
	return nil
}

func LoadSphere(node *yaml.Node, world *scene.Scene, matMap MaterialMap) error {
	var cfg SphereObjectConfig
	err := node.Decode(&cfg)
	if err != nil {
		return err
	}
	if cfg.Position == nil || cfg.Radius == nil {
		return fmt.Errorf("position and radius are required")
	}
	material := matMap.GetWithDefault(cfg.Material)
	obj := &scene.Shading{
		Material: material,
	}
	box := scene.MakeSphere(cfg.Position.X, cfg.Position.Y, cfg.Position.Z, *cfg.Radius)
	box.SetShading(obj)
	box.Add2Scene(world)
	return nil
}

func LoadPlane(node *yaml.Node, world *scene.Scene, matMap MaterialMap) error {
	var cfg PlaneObjectConfig
	err := node.Decode(&cfg)
	if err != nil {
		return err
	}
	if cfg.Size == nil || cfg.Position == nil {
		return fmt.Errorf("size, position and orientation are required")
	}
	plane := scene.MakePlane(
		cfg.Position.X, cfg.Position.Y, cfg.Position.Z, 1, 1)
	if cfg.Orientation != nil {
		plane.Translate(cfg.Position.Negated())
		switch *cfg.Orientation {
			case "+y":
				plane.SwapAxis(1, 2)
			case "-y":
				plane.SwapAxis(1, 2)
				plane.FlipNormals()
			case "+x":
				plane.SwapAxis(0, 2)
			case "-x":
				plane.SwapAxis(0, 2)
				plane.FlipNormals()
			case "+z":
				plane.FlipNormals()
			case "-z":
			default:
				return fmt.Errorf("unknown plane orientation %q", *cfg.Orientation)
		}
		plane.Translate(cfg.Position.Vec3)
		switch (*cfg.Orientation)[1] {
			case 'y':
				plane.Scale(geo.Vec3{cfg.Size[0], 1, cfg.Size[1]}, cfg.Position.Vec3)
			case 'x':
				plane.Scale(geo.Vec3{1, cfg.Size[0], cfg.Size[1]}, cfg.Position.Vec3)
			case 'z':
				plane.Scale(geo.Vec3{cfg.Size[0], cfg.Size[1], 1}, cfg.Position.Vec3)
		}
	}
	material := matMap.GetWithDefault(cfg.Material)
	obj := &scene.Shading{
		Material: material,
	}
	if cfg.Glow != nil {
		obj.Glow = spectra.NewRGBSpectr(cfg.Glow.X, cfg.Glow.Y, cfg.Glow.Z)
	}
	plane.SetShading(obj)
	plane.Add2Scene(world)
	return nil
}

func LoadWeighedSumMaterial(node *yaml.Node) (mat scene.Material, err error) {
	var cfg WeighedSumMaterialConfig
	err = node.Decode(&cfg)
	if err != nil {
		return nil, err
	}
	if cfg.Materials == nil {
		return nil, fmt.Errorf("materials list required")
	}
	if cfg.Weights == nil {
		cfg.Weights = make([]float32, len(cfg.Materials))
		for i := range cfg.Weights {
			cfg.Weights[i] = 1
		}
	}
	if len(cfg.Weights) != len(cfg.Materials) {
		return nil, fmt.Errorf("materials and weights don't match")
	}
	materials := make([]scene.Material, len(cfg.Materials))
	for i, m := range cfg.Materials {
		materials[i], err = LoadMaterial(&m)
		if err != nil {
			return nil, fmt.Errorf("load material %d: %v", i, err)
		}
	}

	return scene.NewWeighedSumMaterial(materials, cfg.Weights), nil
}

func LoadMatteMaterial(node *yaml.Node) (scene.Material, error) {
	var cfg MatteMaterialConfig
	err := node.Decode(&cfg)
	if err != nil {
		return nil, err
	}
	if cfg.Texture == nil && cfg.Color == nil {
		return nil, fmt.Errorf("texture or color is required")
	}
	if cfg.Texture != nil {
		txt := img.LoadPng(*cfg.Texture)
		return scene.NewMatteMaterial(txt, cfg.Roughness, cfg.IsTransparent), nil
	} else {
		c := cfg.Color
		return scene.New1ColorMatteMaterial(
			c.X, c.Y, c.Z, cfg.Roughness, cfg.IsTransparent), nil
	}
}

func LoadBlendMapMaterial(node *yaml.Node) (mat scene.Material, err error) {
	var cfg BlendMapMaterialConfig
	err = node.Decode(&cfg)
	if err != nil {
		return nil, err
	}
	if cfg.Map == "" || cfg.Black.Kind == 0 || cfg.White.Kind == 0 {
		return nil, fmt.Errorf("'map', 'black' and 'white' params are required")
	}

	black, err := LoadMaterial(&cfg.Black)
	if err != nil {
		return nil, fmt.Errorf("black material: %v", err)
	}
	white, err := LoadMaterial(&cfg.White)
	if err != nil {
		return nil, fmt.Errorf("white material: %v", err)
	}
	return scene.NewBlendMapMaterial(black, white, img.LoadPng(cfg.Map)), nil
}

func linearSpectr(a, b float32) (spectra.Spectr) {
	table := spectra.NewSpectrTable()
	for wave := float32(300); wave < float32(885); wave++ {
		power := a + (b - a)*(wave - 300)/(885 - 300)
		table.AppendSample(wave, power)
	}
	return table.MakeRGBSpectr()
}

func LoadMetalMaterial(node *yaml.Node) (scene.Material, error) {
	var cfg MetalMaterialConfig
	err := node.Decode(&cfg)
	if err != nil {
		return nil, err
	}
	var k, eta spectra.Spectr
	if cfg.Color == nil {
		// aluminum approximation
		eta = linearSpectr(0.273375, 2.467289)
		k = linearSpectr(3.59375, 9.98594)
	} else {
		fresnel := func(cosIncidence float32) spectra.Spectr {
			return spectra.NewRGBSpectr(1, 1, 1)
		}
		return scene.NewMicrofacetMaterial(
			spectra.NewRGBSpectr(0, 0, 0),
			cfg.Color.ToSpectr(),
			1.1,
			cfg.Roughness,
			fresnel,
		), nil
	}
	return scene.NewMetalMaterial(eta, k, cfg.Roughness), nil
}

func LoadMaterial(node *yaml.Node) (mat scene.Material, err error) {
	typ, err := DecodeType(node)
	if err != nil {
		return nil, fmt.Errorf("parse type: %v", err)
	}
	var material scene.Material
	switch typ {
		case "matte":
			material, err = LoadMatteMaterial(node)
		case "metal":
			material, err = LoadMetalMaterial(node)
		case "dielectric":
		case "glass":
			material, err = LoadDielectricMaterial(node)
		case "layer":
			material, err = LoadLayerMaterial(node)
		case "blend_map":
			material, err = LoadBlendMapMaterial(node)
		case "weighed_sum":
			material, err = LoadWeighedSumMaterial(node)
		default:
			err = fmt.Errorf("unknown material type %q", typ)
	}
	return material, err
}

func LoadLayerMaterial(node *yaml.Node) (mat scene.Material, err error) {
	var cfg LayerMaterialConfig
	err = node.Decode(&cfg)
	if err != nil {
		return nil, err
	}
	var base scene.Material
	if cfg.Base.Kind == 0 {
		return nil, fmt.Errorf("'base' param is required")
	} else {
		base, err = LoadMaterial(&cfg.Base)
		if err != nil {
			return nil, fmt.Errorf("base material: %v", err)
		}
	}
	if cfg.Eta == nil {
		cfg.Eta = ptrFloat(1.5)
	}

	return scene.NewLayeredMaterial(base, *cfg.Eta), nil
}

func ptrFloat(x float32) *float32 {
	return &x
}

func LoadDielectricMaterial(node *yaml.Node) (scene.Material, error) {
	var cfg DielectricMaterialConfig
	err := node.Decode(&cfg)
	if err != nil {
		return nil, err
	}
	replaceZeroWithDefaults(&cfg, DielectricMaterialConfig{
		Color: &VectorConfig{geo.Vec3{1, 1, 1}},
		ReflectionColor: &VectorConfig{geo.Vec3{1, 1, 1}},
		Eta: ptrFloat(1.5),
		Roughness: ptrFloat(0),
	})
	mtl := scene.NewDielectricMaterial(
		cfg.Color.ToSpectr(),
		cfg.ReflectionColor.ToSpectr(),
		*cfg.Eta,
		*cfg.Roughness,
	)
	return mtl, nil
}

func LoadPerspectiveCamera(node *yaml.Node) (cameras.Camera, error) {
	var cfg PerspectiveCameraConfig
	err := node.Decode(&cfg)
	if err != nil {
		return nil, err
	}
	if cfg.Position == nil || cfg.Target == nil {
		return nil, fmt.Errorf("position and target are required")
	}
	position := cfg.Position.Vec3
	target := cfg.Target.Vec3
	var zoom float32 = 1
	var fov float32 = 0.27*2
	if cfg.Zoom != nil {
		zoom = *cfg.Zoom
	}
	if cfg.Fov != nil {
		fov = (*cfg.Fov) * math.Pi / 180
	}
	cam := cameras.NewPerspectiveCamera(position, target.Sub(position), fov, zoom)
	return cam, nil
}

func LoadOrthoCamera(node *yaml.Node) (cameras.Camera, error) {
	var cfg PerspectiveCameraConfig
	err := node.Decode(&cfg)
	if err != nil {
		return nil, err
	}
	if cfg.Position == nil || cfg.Target == nil {
		return nil, fmt.Errorf("position and target are required")
	}
	position := cfg.Position.Vec3
	target := cfg.Target.Vec3
	var zoom float32 = 1
	if cfg.Zoom != nil {
		zoom = *cfg.Zoom
	}
	cam := cameras.NewOrthoCamera(position, target.Sub(position), zoom)
	return cam, nil
}

func LoadDirectionalLight(node *yaml.Node, world *scene.Scene) error {
	var cfg DirectionalLightConfig
	err := node.Decode(&cfg)
	if err != nil {
		return err
	}
	if cfg.Direction == nil {
		return fmt.Errorf("direction required")
	}
	color := geo.Vec3{1, 1, 1}
	if cfg.Color != nil {
		color = cfg.Color.Vec3
	}
	light := scene.NewDirectionLight(cfg.Direction.Vec3,
		&spectra.RGBSpectr{color.X, color.Y, color.Z}, 10)
	world.AddLight(light)
	return nil
}

func LoadInfiniteAreaLight(node *yaml.Node, world *scene.Scene) error {
	var cfg InfiniteAreaLightConfig
	err := node.Decode(&cfg)
	if err != nil {
		return err
	}
	if cfg.Texture == nil {
		return fmt.Errorf("texture required")
	}
	if cfg.Scale == nil {
		var one float32 = 1
		cfg.Scale = &one
	}
	light := scene.NewInfiniteAreaLight(img.LoadPng(*cfg.Texture), *cfg.Scale)
	if cfg.Scale != nil {
		light.Scale = *cfg.Scale
	}
	if cfg.Direction != nil {
		light.Direction = *cfg.Direction*math.Pi/180
	}
	world.AddLight(light)
	return nil
}

func replaceZeroWithDefaults(s interface{}, defaults interface{}) {
	elem := reflect.ValueOf(s).Elem()
	defVal := reflect.ValueOf(defaults)
	for i := 0; i < elem.NumField(); i++ {
		f := elem.Field(i)
		if f.IsZero() {
			f.Set(defVal.Field(i))
		}
	}
}

func Load(path string, world *scene.Scene) (*SceneConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("fail to open %q", path))
	}
	decoder := yaml.NewDecoder(file)
	var conf SceneConfigYaml
	err = decoder.Decode(&conf)
	if err != nil {
		panic(fmt.Sprintf("decode scene json: %v", err))
	}
	options := conf.Options
	matMap := MaterialMap{
		Map: make(map[string]scene.Material),
		Default: scene.New1ColorMatteMaterial(0.3, 0.6, 1, 0, false),
	}
	for name, node := range conf.Materials {
		material, err := LoadMaterial(&node)
		if err != nil {
			return nil, fmt.Errorf("parse material %q: %v", name, err)
		}
		matMap.Map[name] = material
	}

	type KV struct { k string; v yaml.Node }
	// achtung: very important
	// if list of objects is not sorted, it can lead to non-deterministic scene rendering.
	// found out the hard way :)
	objectList := make([]KV, 0, len(conf.Objects))
	for name, node := range conf.Objects {
		objectList = append(objectList, KV{name, node})
	}
	sort.Slice(objectList, func(i, j int) bool { return objectList[i].k < objectList[j].k })
	for _, kv := range objectList {
		name := kv.k
		node := kv.v
		typ, err := DecodeType(&node)
		if err != nil {
			return nil, fmt.Errorf("parse object %q type: %v", name, err)
		}
		switch typ {
			case "box":
				err = LoadBox(&node, world, matMap)
			case "sphere":
				err = LoadSphere(&node, world, matMap)
			case "obj":
				err = LoadObj(&node, world, matMap)
			case "plane":
				err = LoadPlane(&node, world, matMap)
				
			default:
				err = fmt.Errorf("unknown object type %q", typ)
		}
		if err != nil {
			return nil, fmt.Errorf("parse object %q: %v", name, err)
		}
		var obj ObjectConfig
		err = node.Decode(&obj)
		if err != nil {
			panic("aaa")
		}
	}
	for name, node := range conf.Lights {
		typ, err := DecodeType(&node)
		if err != nil {
			return nil, fmt.Errorf("parse light %q type: %v", name, err)
		}
		switch typ {
			case "directional":
				err = LoadDirectionalLight(&node, world)
			case "infinite":
				err = LoadInfiniteAreaLight(&node, world)
				
			default:
				err = fmt.Errorf("unknown light type %q", typ)
		}
		if err != nil {
			return nil, fmt.Errorf("parse light %q: %v", name, err)
		}
	}
	var camera cameras.Camera
	for name, node := range conf.Cameras {
		typ, err := DecodeType(&node)
		if err != nil {
			return nil, fmt.Errorf("parse camera %q type: %v", name, err)
		}
		switch typ {
			case "perspective":
				cam, err := LoadPerspectiveCamera(&node)
				if err != nil {
					break
				}
				if name == conf.ActiveCamera {
					camera = cam
				}
				
			case "orthographic":
				cam, err := LoadOrthoCamera(&node)
				if err != nil {
					break
				}
				if name == conf.ActiveCamera {
					camera = cam
				}
				
			default:
				err = fmt.Errorf("unknown camera type %q", typ)
		}
		if err != nil {
			return nil, fmt.Errorf("parse camera %q: %v", name, err)
		}
	}
	if conf.Profile == "" {
		conf.Profile = "main"
	}
	profile := conf.Profiles[conf.Profile]
	replaceZeroWithDefaults(&profile, ProfileConfig{
		Width: 400,
		Height: 400,
		PixelSamples: 10,
		SaveInterval: 600,
	})
	replaceZeroWithDefaults(&options, Options{
		Goroutines: 4,
		Outfile: "out.png",
		Region: [4]int{0, 0, profile.Width, profile.Height},
	})
	options.Profile = profile

	switch conf.Accelerator {
		case "bvh":
			tree := scene.MakeBVH(world.Shapes)
			world.Accelerator = tree
		case "":
		default:
			return nil, fmt.Errorf("unknown accelerator %q", conf.Accelerator)
	}

	world.Preprocess()

	ret := SceneConfig{
		Camera: camera,
		Options: &options,
	}
	if profile.Tracer.Kind == 0 {
		ret.Tracer = tracers.NewPathTracer(0, 0)
	} else {
		tracerType, err := DecodeType(&profile.Tracer)
		if err != nil {
			return nil, fmt.Errorf("parse tracer type: %v", err)
		}
		switch tracerType {
			case "path":
				var cfg PathTracerConfig
				err := profile.Tracer.Decode(&cfg)
				if err != nil {
					return nil, fmt.Errorf("load path tracer config: %s", err)
				}
				ret.Tracer = tracers.NewPathTracer(cfg.MinDepth, cfg.TerminationProb)
			case "direct":
				ret.Tracer = tracers.NewDirectTracer()
			case "ftl":
				var cfg FTLTracerConfig
				err := profile.Tracer.Decode(&cfg)
				if err != nil {
					return nil, fmt.Errorf("load ftl tracer config: %s", err)
				}
				if cfg.Fps == 0 {
					cfg.Fps = 1
				}
				if cfg.OutfilePattern == "" {
					return nil, fmt.Errorf("load ftl tracer config: outfile_pattern required")
				}
				if cfg.NFrames == 0 {
					return nil, fmt.Errorf("load ftl tracer config: n_frames required")
				}
				ret.FTLTracer = tracers.NewFTLTracer(
					cfg.MinDepth,
					cfg.TerminationProb,
					cfg.NFrames,
					cfg.Fps,
				)
				ret.FTLTracer.TimeOffset = cfg.TimeOffset
				ret.FTLTracer.SkipFirstSegment = cfg.SkipFirstSegment
				if cfg.LightDuration != 0 {
					ret.FTLTracer.LightDuration = cfg.LightDuration
				}
				options.Outfile = cfg.OutfilePattern
			case "dum":
			default:
				return nil, fmt.Errorf("unknown tracer: %q", tracerType)
		}
	}
	return &ret, nil
}

func Noop() {
	img.Noop()
}
