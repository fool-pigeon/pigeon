package service

import (
	"errors"
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

type NodeMode uint32

const (
	DirMode     NodeMode = 1 << iota // Directory: is a directory
	ConfigDir                        // ConfigDir: the .pg/ directory
	NormalDir                        // NormalDir: redundancy
	FileMode                         // File: is a file
	Request                          // Req: is a request file
	HttpReq                          // HTTP:[httpx.ReqData]
	HttpGet                          // GET: http GET Method
	HttpHead                         // HEAD: http HEAD Mehtod
	HttpPost                         // POST: http POST Method
	HttpPut                          // PUT: http PUT Method
	HttpPatch                        // PATCH: http PATCH Method
	HttpDelete                       // DELETE: http DELETE Method
	HttpConnect                      // CONNECT: http CONNECT Method
	HttpOptions                      // OPTIONS: http OPTIONS Method
	HttpTrace                        // TRACE: http TRACE Method
)

func (m NodeMode) HasRole(role NodeMode) bool {
	return m&role != 0
}
func (m NodeMode) IsDir() bool {
	return m&DirMode != 0
}

type Node struct {
	Mode NodeMode `json:"mode"`
	Path string   `json:"path"`

	Name string `json:"name"`

	// for cache view
	SubCache []*Node `json:"sub"`

	sub map[string]*Node
	// parent node
	pre *Node
	// for lazy load & render
	loaded bool

	mu sync.RWMutex
}

func NewNode(mode NodeMode, path string, name string, pre ...*Node) *Node {
	if len(pre) > 0 {
		return &Node{Path: path, Name: name, Mode: mode, pre: pre[0]}
	}
	return &Node{Path: path, Name: name, Mode: mode}
}
func NewDirNode(path string, name string, pre ...*Node) *Node {
	m := DirMode
	if name == fsx.RootTagDirName {
		m |= ConfigDir
	} else {
		m |= NormalDir
	}
	return NewNode(m, path, name, pre...)
}

func (n *Node) GetMode() NodeMode {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.Mode
}
func (n *Node) SetMode(t NodeMode) *Node {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Mode = t
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

// first time to load the dir's subs
func (n *Node) load() error {
	if !n.GetMode().IsDir() {
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
			if _, ok := subs[path]; ok {
				return
			}
			subs[path] = NewNode(getFileMode(b), path, d.Name(), n)
		})
	}
	wg.Wait()
	n.loaded = true
	n.sub = subs
	return nil
}

// reload dir's subs with the minimum modification
func (n *Node) reload() error {
	if !n.GetMode().IsDir() {
		return nil
	}
	if !n.Loaded() {
		return nil
	}
	parent := n.GetPath()
	ds, err := os.ReadDir(parent)
	if err != nil {
		return err
	}

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)
	subs := make(map[string]*Node)
	for _, d := range ds {
		path := filepath.Join(parent, d.Name())

		wg.Go(func() {
			mu.Lock()
			if node, ok := n.Sub(path); ok {
				defer mu.Unlock()
				subs[path] = node
				return
			}

			if d.IsDir() {
				defer mu.Unlock()
				subs[path] = NewDirNode(path, d.Name(), n)
				return
			}
			mu.Unlock()

			b, _ := fsx.ReadFile(path)
			mu.Lock()
			defer mu.Unlock()
			if _, ok := subs[path]; ok {
				return
			}
			subs[path] = NewNode(getFileMode(b), path, d.Name(), n)
		})
	}
	wg.Wait()
	n.mu.Lock()
	defer n.mu.Unlock()
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
	subs = slices.AppendSeq(subs, maps.Values(n.sub))
	n.mu.RUnlock()

	slices.SortFunc(subs, func(a *Node, b *Node) int {
		if m1, m2 := a.GetMode(), b.GetMode(); m1 != m2 {
			return int(m1) - int(m2)
		} else {
			return strings.Compare(a.GetName(), b.GetName())
		}
	})

	n.SetSubsCache(subs)
}

// parse FSType from file content
func getFileMode(b []byte) NodeMode {
	res := FileMode
	// parse if it is a request file
	var reqMeta request.ReqMeta
	if err := toml.Unmarshal(b, &reqMeta); err == nil {
		// add Request flag
		res |= Request
		switch reqMeta.ReqProtocol() {
		case request.P_HTTP:
			// add HTTP flag
			var httpReq httpx.ReqData
			toml.Unmarshal(b, &httpReq)
			res |= parseModeFromHttpReq(&httpReq)
		}
		return res
	}

	return res
}
func parseModeFromHttpReq(req *httpx.ReqData) NodeMode {
	res := HttpReq
	switch req.Method {
	case httpx.MethodGet:
		res |= HttpGet
	case httpx.MethodPost:
		res |= HttpPost
	case httpx.MethodPut:
		res |= HttpPut
	case httpx.MethodDelete:
		res |= HttpDelete
	case httpx.MethodPatch:
		res |= HttpPatch
	case httpx.MethodHead:
		res |= HttpHead
	case httpx.MethodOptions:
		res |= HttpOptions
	case httpx.MethodTrace:
		res |= HttpTrace
	case httpx.MethodConnect:
		res |= HttpConnect
	}
	return res
}

var (
	ErrNilRoot      = errors.New("root is nil")
	ErrNodeNotExist = errors.New("node not exist")
	ErrNodeNilPre   = errors.New("node's pre is nil")
)

type FileExploreService struct {
	mu   sync.RWMutex
	root *Node
	em   *application.EventManager
}

func NewFileExploreService(em *application.EventManager) *FileExploreService {
	return &FileExploreService{root: &Node{Mode: DirMode}, em: em}
}

// ServiceStartup implements [application.ServiceStartup].
// func (f *FileExploreService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
// 	f.mu.Lock()
// 	defer f.mu.Unlock()
// 	f.root = &Node{Type: DirType}
// 	return nil
// }

// Parse the root dir paths in the directory
func (f *FileExploreService) ParseRootPaths(dir string) []string {
	rootPaths := make([]string, 0)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var parseDir func(string)
	parseDir = func(path string) {
		info, err := os.Lstat(path)
		// cause error or not a dir
		if err != nil || !info.IsDir() {
			return
		}

		// has the .pg/ means this is a root dir
		if tagInfo, err := os.Lstat(filepath.Join(path, fsx.RootTagDirName)); err == nil && tagInfo.IsDir() {
			mu.Lock()
			rootPaths = append(rootPaths, path)
			mu.Unlock()
		}

		// parse sub dirs
		entries, _ := os.ReadDir(path)
		for _, entry := range entries {
			if entry.IsDir() {
				wg.Go(func() { parseDir(filepath.Join(path, entry.Name())) })
			}
		}

	}
	wg.Go(func() { parseDir(filepath.Clean(dir)) })
	wg.Wait()
	return rootPaths
}

// add workspace's root dir
func (f *FileExploreService) AddRoot(paths []string) ([]*Node, error) {
	f.mu.RLock()
	if f.root == nil {
		return nil, ErrNilRoot
	}
	f.mu.RUnlock()

	for _, path := range paths {
		path = filepath.Clean(path)
		// has existed
		if f.root.SubContain(path) {
			continue
		}
		if info, err := os.Lstat(filepath.Join(path, fsx.RootTagDirName)); !info.IsDir() || err != nil {
			continue
		}
		info, err := os.Lstat(path)
		if err != nil || !info.IsDir() {
			continue
		}
		f.root.SetSub(path, NewDirNode(path, info.Name(), f.root))
	}

	f.root.SetLoaded(true)
	f.root.updateCache()
	return f.root.SubsCache(), nil
}

// load the dir
func (f *FileExploreService) LoadSubs(dir string) error {
	node, err := f.getNodeExact(dir)
	if err != nil {
		return err
	}
	return node.load()
}

// load the dir and return subsCache
func (f *FileExploreService) LoadSubsCached(dir string) ([]*Node, error) {
	node, err := f.getNodeExact(dir)
	if err != nil {
		return nil, err
	}
	if err := node.load(); err != nil {
		return nil, err
	}
	node.updateCache()
	return node.SubsCache(), nil
}

// reload dir
func (f *FileExploreService) ReloadSubs(dir string) error {
	node, err := f.getNodeExact(dir)
	if err != nil {
		return err
	}
	return node.reload()
}

// reload dir and return subsCache
func (f *FileExploreService) ReloadSubsCached(dir string) ([]*Node, error) {
	node, err := f.getNodeExact(dir)
	if err != nil {
		return nil, err
	}
	if err := node.reload(); err != nil {
		return nil, err
	}
	node.updateCache()
	return node.SubsCache(), nil
}

// return current [Node.SubsCache] after called [FileExploreService.UpdateView]
func (f *FileExploreService) ReloadRootView() ([]*Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.root == nil {
		return nil, ErrNilRoot
	}

	var wg sync.WaitGroup
	nodes := []*Node{f.root}
	q := []*Node{f.root}
	for len(q) > 0 {
		for _, node := range q[0].Subs() {
			if !node.Loaded() {
				continue
			}
			if info, err := os.Lstat(node.GetPath()); err == nil && info.IsDir() {
				nodes = append(nodes, node)
				q = append(q, node)
			}

		}
		q = q[1:]
	}
	for _, node := range nodes {
		wg.Go(func() {
			node.reload()
			node.updateCache()
		})
	}
	wg.Wait()
	return f.root.SubsCache(), nil
}

// Delete Node
func (f *FileExploreService) DeleteNode(path string) error {
	node, err := f.getNodeExact(path)
	if err != nil {
		return err
	}

	path = node.GetPath()
	pre := node.Pre()
	if pre == nil {
		return ErrNodeNilPre
	}
	if err := os.RemoveAll(path); err != nil {
		return err
	}

	pre.DelSub(path)

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
