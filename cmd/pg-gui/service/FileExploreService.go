package service

import (
	"context"
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/fool-pigeon/pigeon/internal/request"
	"github.com/fool-pigeon/pigeon/internal/request/httpx"
	"github.com/pelletier/go-toml/v2"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type FSType int

const (
	TypeUnkown FSType = iota
	DirType
	FileType
)

// identify file content type
type FSTag int

const (
	Unkown FSTag = iota

	HttpGet
	HttpHead
	HttpPost
	HttpPut
	HttpPatch
	HttpDelete
	HttpConnect
	HttpOptions
	HttpTrace
)

type Node struct {
	Type FSType `json:"type"`
	Path string `json:"path"`

	Name string `json:"name"`
	Tag  FSTag  `json:"tag"`

	// for cache view
	SubCache []*Node `json:"sub"`

	sub map[string]*Node
	// parent node
	pre *Node
	// for lazy load & render
	loaded bool

	mu sync.RWMutex
}

func NewDirNode(path string, name string, pre ...*Node) *Node {
	if len(pre) > 0 {
		return &Node{Type: DirType, Path: path, Name: name, pre: pre[0]}
	}
	return &Node{Type: DirType, Path: path, Name: name}
}
func NewFileNode(path string, name string, tag FSTag, pre ...*Node) *Node {
	if len(pre) > 0 {
		return &Node{Type: FileType, Path: path, Name: name, Tag: tag, pre: pre[0]}
	}
	return &Node{Type: FileType, Path: path, Name: name, Tag: tag}
}
func (n *Node) GetType() FSType {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.Type
}
func (n *Node) SetType(t FSType) *Node {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Type = t
	return n
}
func (n *Node) GetPath() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.Path
}
func (n *Node) SetPath(path string) *Node {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Path = path
	return n
}
func (n *Node) GetName() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.Name
}
func (n *Node) SetName(name string) *Node {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Name = name
	return n
}
func (n *Node) GetTag() FSTag {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.Tag
}
func (n *Node) SetTag(tag FSTag) *Node {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Tag = tag
	return n
}
func (n *Node) Loaded() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.loaded
}
func (n *Node) SetLoaded(loaded bool) *Node {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.loaded = loaded
	return n
}
func (n *Node) Pre() *Node {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.pre
}
func (n *Node) SetPre(pre *Node) *Node {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.pre = pre
	return n
}
func (n *Node) Subs() []*Node {
	n.mu.RLock()
	defer n.mu.RUnlock()
	subs := make([]*Node, 0, len(n.sub))
	return slices.AppendSeq(subs, maps.Values(n.sub))
}
func (n *Node) Sub(path string) (*Node, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	node, ok := n.sub[path]
	return node, ok
}
func (n *Node) SubContain(path string) bool {
	_, ok := n.Sub(path)
	return ok
}
func (n *Node) SubLen() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.sub)
}
func (n *Node) SetSubs(subs map[string]*Node) *Node {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.sub = subs
	return n
}
func (n *Node) SetSub(path string, node *Node) *Node {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.sub == nil {
		n.sub = make(map[string]*Node)
	}
	n.sub[path] = node
	return n
}
func (n *Node) SubsCache() []*Node {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.SubCache
}
func (n *Node) SetSubsCache(subs []*Node) *Node {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.SubCache = subs
	return n
}

// Add a node to sub that not nil
func (n *Node) AddSub(node *Node) {
	if node == nil {
		return
	}
	n.SetSub(node.GetPath(), node)
}

// Remove the node from sub whose path is <path>
func (n *Node) DelSub(path string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.sub == nil {
		return
	}
	delete(n.sub, path)
}
func (n *Node) load() error {
	if n.GetType() != DirType {
		return nil
	}
	parent := n.GetPath()
	ds, err := os.ReadDir(parent)
	if err != nil {
		return err
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.loaded {
		return nil
	}

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)
	subs := make(map[string]*Node)
	for _, d := range ds {
		path := filepath.Join(parent, d.Name())
		if d.IsDir() {
			subs[path] = NewDirNode(path, d.Name(), n)
			continue
		}
		wg.Go(func() {
			b, _ := fsx.ReadFile(path)
			mu.Lock()
			defer mu.Unlock()
			subs[path] = NewFileNode(path, d.Name(), getTag(b), n)
		})
	}
	wg.Wait()
	n.loaded = true
	n.sub = subs
	return nil
}
func (n *Node) updateCache() {
	if !n.Loaded() {
		return
	}

	if n.SubLen() <= 0 {
		n.SetSubsCache(nil)
		return
	}

	n.mu.RLock()
	subs := make([]*Node, 0, len(n.sub))
	for _, v := range n.sub {
		subs = append(subs, v)
	}
	n.mu.RUnlock()

	slices.SortFunc(subs, func(a *Node, b *Node) int {
		if type1, type2 := a.GetType(), b.GetType(); type1 != type2 {
			return int(type1) - int(type2)
		} else if tag1, tag2 := a.GetTag(), b.GetTag(); tag1 != tag2 {
			return int(tag1) - int(tag2)
		} else {
			return strings.Compare(a.GetName(), b.GetName())
		}
	})

	n.SetSubsCache(subs)
}

// switch tag
func getTag(b []byte) FSTag {
	var reqMeta request.ReqMeta
	toml.Unmarshal(b, &reqMeta)
	switch reqMeta.ReqProtocol() {
	case request.P_HTTP:
		var httpReq httpx.ReqData
		toml.Unmarshal(b, &httpReq)
		switch httpReq.Method {
		case httpx.MethodGet:
			return HttpGet
		case httpx.MethodPost:
			return HttpPost
		case httpx.MethodPut:
			return HttpPut
		case httpx.MethodDelete:
			return HttpDelete
		case httpx.MethodPatch:
			return HttpPatch
		case httpx.MethodHead:
			return HttpHead
		case httpx.MethodOptions:
			return HttpOptions
		case httpx.MethodTrace:
			return HttpTrace
		case httpx.MethodConnect:
			return HttpConnect
		}
	}
	return Unkown
}

var (
	ErrNilRoot      = errors.New("root is nil")
	ErrNodeNotExist = errors.New("node not exist")
	ErrNodeNilPre   = errors.New("node's pre is nil")
)

type FileExploreService struct {
	mu   sync.RWMutex
	root *Node
}

// ServiceStartup implements [application.ServiceStartup].
func (f *FileExploreService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.root = &Node{Type: DirType}
	return nil
}

// path is a directory path
func (f *FileExploreService) LoadRoot(dir string) ([]*Node, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.root == nil {
		return nil, ErrNilRoot
	}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		// cause error or not a directory
		if err != nil || !d.IsDir() {
			return nil
		}

		path = filepath.Clean(path)
		// has existed
		if f.root.SubContain(path) {
			return filepath.SkipDir
		}
		// not has root tag or cause error
		if isExist, err := fsx.IsDirExist(filepath.Join(path, fsx.RootTagDirName)); !isExist || err != nil {
			return nil
		}

		f.root.SetSub(path, NewDirNode(path, d.Name(), f.root))
		return nil
	})
	if err != nil {
		return nil, err
	}
	f.root.SetLoaded(true)
	f.root.updateCache()
	return f.root.SubsCache(), nil
}

// load the dir
func (f *FileExploreService) Load(dir string) error {
	node, err := f.getNode(filepath.Clean(dir))
	if err != nil {
		return err
	}

	return node.load()
}

// return current [Node.SubsCache] after called [FileExploreService.UpdateView]
func (f *FileExploreService) UpdateRoot() ([]*Node, error) {
	if err := f.UpdateView(); err != nil {
		return nil, err
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.root.SubsCache(), nil
}

// Update the subCache for all loaded node
func (f *FileExploreService) UpdateView() error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.root == nil {
		return ErrNilRoot
	}

	var wg sync.WaitGroup
	nodes := []*Node{f.root}
	q := []*Node{f.root}
	for len(q) > 0 {
		if front := q[0]; front.Loaded() {
			subs := front.Subs()
			nodes = append(nodes, subs...)
			q = append(q, subs...)
		}
		q = q[1:]
	}
	for _, node := range nodes {
		wg.Go(func() {
			node.updateCache()
		})
	}
	wg.Wait()
	return nil
}

// Delete Node
func (f *FileExploreService) DeleteNode(path string) error {
	node, err := f.getNodeExact(path)
	if err != nil {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	path = node.GetPath()
	pre := node.Pre()
	if pre == nil {
		return ErrNodeNilPre
	}
	pre.DelSub(path)
	os.RemoveAll(path)

	return nil
}

// Add Node
func (f *FileExploreService) AddNode(parent string, node *Node) error {
	pre, err := f.getNodeExact(parent)
	if err != nil {
		return err
	}
	node.SetPath(filepath.Join(pre.GetPath(), node.GetName())).SetPre(pre)
	pre.AddSub(node)
	return nil
}

// get the node that matches path as closely as possible
func (f *FileExploreService) getNode(path string) (*Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.root == nil {
		return nil, ErrNilRoot
	}
	// Parse paths form OS' root to current path
	paths := []string{}
	for p := path; p != filepath.Dir(p); p = filepath.Dir(p) {
		paths = append(paths, p)
	}
	slices.Reverse(paths)
	// Find the node that matches path as closely as possible
	node := f.root
	for _, p := range paths {
		if next, ok := node.Sub(p); ok {
			node = next
		}
	}
	// not find node
	if f.root == node {
		return nil, ErrNodeNotExist
	}
	return node, nil
}

// get node whose path equals <path>
func (f *FileExploreService) getNodeExact(path string) (*Node, error) {
	node, err := f.getNode(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	if node.GetPath() != filepath.Clean(path) {
		return nil, ErrNodeNotExist
	}
	return node, nil
}
