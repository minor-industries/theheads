package scene

import (
	"bytes"
	"fmt"
	geom2 "github.com/minor-industries/platform/common/geom"
	"github.com/minor-industries/platform/common/timed_reset"
	"github.com/minor-industries/theheads/common/schema"
	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"
)

const (
	defaultCameraSensitivity = 0.2
)

type Scene struct {
	Stands        []*Stand
	Scale         int
	Translate     Translate
	Scenes        []string
	StartupScenes []string

	// TODO: don't hang these config values off of here
	CameraSensitivity float64

	HeadMap map[string]*Head `toml:"-"`
	Heads   []*Head

	CameraMap map[string]*Camera `toml:"-"`
	Cameras   []*Camera

	Texts []*Text `toml:"-"`
}

type Pos struct {
	X float64
	Y float64
}

type Camera struct {
	Description string
	Fov         float64
	Name        string
	Pos         Pos
	Rot         float64

	M     geom2.Mat `toml:"-"`
	Stand *Stand    `toml:"-"`
	Path  []string  `toml:"-"`
}

func (c Camera) URI() string {
	return fmt.Sprintf("camera://%s/%s", strings.Join(c.Path, "/"), c.Name)
}

type Head struct {
	Name string

	Pos     Pos
	Rot     float64
	Virtual bool

	Path  []string  `toml:"-"`
	M     geom2.Mat `toml:"-"`
	MInv  geom2.Mat `toml:"-"`
	Stand *Stand    `toml:"-"`

	fearful *timed_reset.Bool
}

func (h *Head) URI() string {
	return fmt.Sprintf("head://%s/%s", strings.Join(h.Path, "/"), h.Name)
}

func (h *Head) LedsURI() string {
	return fmt.Sprintf("leds://%s/%s", strings.Join(h.Path, "/"), h.Name)
}

func (h *Head) CameraURI() string {
	return fmt.Sprintf(
		"camera://%s/%s",
		strings.Join(h.Path, "/"),
		strings.Replace(h.Name, "head", "camera", -1), // TODO: should actually traverse entity tree
	)
}

func (h *Head) Fearful() bool {
	return h.fearful.Val()
}

func SelectHeads(
	heads map[string]*Head,
	predicate func(i int, h *Head) bool,
) (result []*Head) {
	i := 0
	for _, head := range heads {
		if predicate(i, head) {
			result = append(result, head)
		}
		i++
	}
	return
}

func ShuffledHeads(heads map[string]*Head) (result []*Head) {
	for _, head := range heads {
		result = append(result, head)
	}
	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})
	return
}

func (s *Scene) HeadURIs() []string {
	var result []string
	for _, head := range s.Heads {
		result = append(result, head.URI())
	}
	return result
}

func (s *Scene) OnFaceDetected(msg *schema.FaceDetected) error {
	heads, err := s.findHeadsForCamera(msg.CameraName)
	if err != nil {
		return errors.Wrap(err, "find heads for camera")
	}
	for _, head := range heads {
		duration := time.Duration(20+rand.Intn(10)) * time.Second
		head.fearful.SetFor(duration)
	}
	return nil
}

func (s *Scene) findHeadsForCamera(name string) ([]*Head, error) {
	camera, ok := s.CameraMap[name]
	if !ok {
		return nil, errors.New("unknown camera")
	}
	stand := camera.Stand
	if stand == nil {
		return nil, nil
	}
	return stand.Heads, nil
}

func (s *Scene) ClearFearful() {
	for _, head := range s.HeadMap {
		head.fearful.Clear()
	}
}

type Stand struct {
	Name string

	CameraNames []string
	HeadNames   []string

	Pos Pos
	Rot float64

	Disabled bool

	M geom2.Mat `toml:"-"`

	Cameras []*Camera `toml:"-"`
	Heads   []*Head   `toml:"-"`

	CameraMap map[string]*Camera `toml:"-"`
	HeadMap   map[string]*Head   `toml:"-"`
}
type Translate struct {
	X int
	Y int
}

func (h *Head) GlobalPos() geom2.Vec {
	t2 := h.Stand
	t1 := t2.M
	t3 := h.M
	t0 := t1.Mul(t3)
	return t0.Translation()
}
func (h *Head) PointAwayFrom(p geom2.Vec) float64 {
	return math.Mod(h.PointTo(p)+180.0+360.0, 360.0)
}

func (h *Head) PointTo(p geom2.Vec) float64 {
	to := h.MInv.MulVec(p)
	theta := math.Atan2(to.Y(), to.X()) * 180 / math.Pi
	return math.Mod(theta+360.0, 360)
}

func getPrefix(prefix string) (map[string][]byte, error) {
	dir, err := ioutil.ReadDir(prefix)
	if err != nil {
		return nil, errors.Wrap(err, "readdir "+prefix)
	}

	result := map[string][]byte{}
	for _, info := range dir {
		filename := path.Join(prefix, info.Name())

		ext := path.Ext(filename)
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			continue
		}

		content, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, errors.Wrap(err, "readdir "+info.Name())
		}

		result[info.Name()] = content
	}

	return result, nil
}

func BuildInstallation(scenePath, sceneName, textSet string) (*Scene, error) {
	fullPath := path.Join(scenePath, sceneName+".toml")

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, errors.Wrap(err, "readfile")
	}

	sc, err := Build(content)
	if err != nil {
		return nil, errors.Wrap(err, "build")
	}

	// Texts
	sc.Texts = LoadTexts(scenePath, textSet)
	return sc, nil
}

func Build(content []byte) (*Scene, error) {
	scene := &Scene{
		CameraMap: map[string]*Camera{},
		HeadMap:   map[string]*Head{},
	}

	d := toml.NewDecoder(bytes.NewBuffer(content))
	d.DisallowUnknownFields()

	err := d.Decode(scene)
	if err != nil {
		return nil, errors.Wrap(err, "decode toml")
	}

	// heads and cameras don't "exist" until they are added to a stand
	definedHeads := map[string]*Head{}
	definedCameras := map[string]*Camera{}

	// CameraMap

	for _, camera := range scene.Cameras {
		camera.M = geom2.ToM(camera.Pos.X, camera.Pos.Y, camera.Rot)
		definedCameras[camera.Name] = camera
	}

	for _, head := range scene.Heads {
		head.fearful = timed_reset.NewBool()
		head.M = geom2.ToM(head.Pos.X, head.Pos.Y, head.Rot)
		definedHeads[head.Name] = head
	}

	for _, stand := range scene.Stands {
		// TODO: enabled/disabled
		stand.CameraMap = map[string]*Camera{}
		stand.HeadMap = map[string]*Head{}
		stand.M = geom2.ToM(stand.Pos.X, stand.Pos.Y, stand.Rot)

		for _, name := range stand.CameraNames {
			camera, ok := definedCameras[name]
			if !ok {
				return nil, errors.New(fmt.Sprintf("%s not found", name))
			}
			camera.Path = []string{stand.Name}
			camera.Stand = stand
			stand.CameraMap[camera.Name] = camera
			stand.Cameras = append(stand.Cameras, camera)
			scene.CameraMap[camera.Name] = camera
			scene.Cameras = append(scene.Cameras, camera)
		}

		for _, name := range stand.HeadNames {
			head, ok := definedHeads[name]
			if !ok {
				return nil, errors.New(fmt.Sprintf("%s not found", name))
			}
			head.Path = []string{stand.Name}
			head.Stand = stand
			head.MInv = head.Stand.M.Mul(head.M).Inv() // hmmmm, we use Stand.M for MInv but not for head.M
			stand.HeadMap[head.Name] = head
			stand.Heads = append(stand.Heads, head)
			scene.HeadMap[head.Name] = head
			scene.Heads = append(scene.Heads, head)
		}
	}

	if scene.CameraSensitivity == 0 {
		scene.CameraSensitivity = defaultCameraSensitivity

	}

	return scene, nil
}
