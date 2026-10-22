// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package service

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/fool-pigeon/pigeon/core"
	"github.com/fool-pigeon/pigeon/fsx"
	"github.com/fool-pigeon/pigeon/protocol"
	"github.com/fool-pigeon/pigeon/uid"
	"github.com/pelletier/go-toml/v2"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type NodeMode uint32

const (
	DirMode      NodeMode = 1 << iota // Directory: is a directory
	RootTag                           // the .pg/ directory
	RootConfFile                      // e.g .pg/conf/conf.toml
	RootRunFile                       // e.g .pg/run/dev.toml
	NormalDir                         // NormalDir: redundancy
	FileMode                          // File: is a file
	ConfFile                          // ConfFile: toml file named .conf

	Request // Req: is a request file
	ReqREST // [protocol.ProtocolREST]
	ReqGQL  // [protocol.ProtocolGQL]
	ReqGRPC // [protocol.ProtocolGRPC]
	ReqWS   // [protocol.ProtocolWS]

	HttpGet     // GET: http GET Method
	HttpPost    // POST: http POST Method
	HttpPut     // PUT: http PUT Method
	HttpDelete  // DELETE: http DELETE Method
	HttpHead    // HEAD: http HEAD Mehtod
	HttpPatch   // PATCH: http PATCH Method
	HttpOptions // OPTIONS: http OPTIONS Method
	HttpConnect // CONNECT: http CONNECT Method
	HttpTrace   // TRACE: http TRACE Method
)

func (m NodeMode) Has(role NodeMode) bool {
	return m&role != 0
}
func (m NodeMode) IsDir() bool {
	return m&DirMode != 0
}

var (
	ErrNodeNotDir  = errors.New("Node is not a directory.")
	ErrNodeDir     = errors.New("Node is a directory.")
	ErrNodeNotFile = errors.New("Node is not a file.")
	ErrNodeFile    = errors.New("Node is a file.")
)

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
	if name == fsx.RootTag {
		m |= RootTag
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
	var req protocol.Request

	if err := toml.Unmarshal(b, &req); err == nil {
		// add Request flag
		res |= Request
		switch req.Protocol {
		case protocol.ProtocolREST, protocol.ProtocolGQL:
			// add HTTP flag
			switch req.Method {
			case protocol.MethodGet:
				res |= HttpGet
			case protocol.MethodPost:
				res |= HttpPost
			case protocol.MethodPut:
				res |= HttpPut
			case protocol.MethodDelete:
				res |= HttpDelete
			case protocol.MethodPatch:
				res |= HttpPatch
			case protocol.MethodHead:
				res |= HttpHead
			case protocol.MethodOptions:
				res |= HttpOptions
			case protocol.MethodTrace:
				res |= HttpTrace
			case protocol.MethodConnect:
				res |= HttpConnect
			}
		}
		return res
	}

	return res
}

// FileExploreService Event
type FesEvent = string

const (
	// Send data:
	// {path:string,subs:[]*Node}
	FesUpdateSubs FesEvent = "file-explore-service:update-subs"
	// Send data:
	// (path:string)
	FesDeletedNode FesEvent = "file-explore-service:deleted-node"
	// Send data:
	// {path:string,subs:[]*Node}
	FesRename FesEvent = "file-explore-service:rename"
	// Send data:
	// {path:string,subs:[]*Node}
	FesReloadRoot FesEvent = "file-explore-service:rename"
)

type FesEvtUpdateSubs struct {
	Path string  `json:"path"`
	Subs []*Node `json:"subs"`
}

var (
	ErrNilRoot      = errors.New("root is nil")
	ErrNodeNotExist = errors.New("node not exist")
	ErrNodeNilPre   = errors.New("node's pre is nil")
)

type FileExploreService struct {
	mu    sync.RWMutex
	root  *Node
	alert *AlertService
	app   *application.App
}

func NewFileExploreService(alertService *AlertService, app *application.App) *FileExploreService {
	return &FileExploreService{root: &Node{Mode: DirMode}, alert: alertService, app: app}
}

// Find the node whose path equals the param.
// returns nil if not found.
func (f *FileExploreService) FindNode(path string) *Node {
	path = filepath.Clean(path)
	q := f.root.Subs()
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		curPath := cur.GetPath()
		if curPath == path {
			return cur
		}
		if IsStrictParent(curPath, path) && cur.Loaded() && cur.SubLen() > 0 {
			rel, _ := filepath.Rel(curPath, path)
			parts := strings.Split(rel, string(filepath.Separator))
			if node, ok := cur.Sub(filepath.Join(curPath, parts[0])); ok {
				q = append(q, node)
			}
		}
	}
	return nil
}

// IsStrictParent reports whether base is a strict parent directory of target.
// base itself is not a parent of base.
// Lexical comparison only; symbolic links are not resolved.
func IsStrictParent(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false // different volumes on Windows, or invalid path
	}
	// rel == "." means base == target → not a parent
	// rel == ".." means target is base's parent → not a parent
	// rel starts with ".." + separator means target escapes base → not a parent
	return rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
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
		if tagInfo, err := os.Lstat(filepath.Join(path, fsx.RootTag)); err == nil && tagInfo.IsDir() {
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
func (f *FileExploreService) AddRoot(paths []string) error {
	f.mu.RLock()
	if f.root == nil {
		return ErrNilRoot
	}
	f.mu.RUnlock()

	for _, path := range paths {
		path = filepath.Clean(path)
		// has existed
		if f.root.SubContain(path) {
			continue
		}
		if info, err := os.Lstat(filepath.Join(path, fsx.RootTag)); !info.IsDir() || err != nil {
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
	f.app.Event.Emit(FesUpdateSubs, FesEvtUpdateSubs{
		Path: f.root.GetPath(),
		Subs: f.root.SubsCache(),
	})
	return nil
}

// load the dir
func (f *FileExploreService) LoadSubs(dir string) error {
	node, err := f.GetNodeExact(dir)
	if err != nil {
		return err
	}
	return node.load()
}

// load the dir and return subsCache
func (f *FileExploreService) LoadSubsCached(dir string) error {
	node, err := f.GetNodeExact(dir)
	if err != nil {
		return err
	}
	if err := node.load(); err != nil {
		return err
	}
	node.updateCache()
	f.app.Event.Emit(FesUpdateSubs, FesEvtUpdateSubs{
		Path: node.GetPath(),
		Subs: node.SubsCache(),
	})
	return nil
}

// reload dir
func (f *FileExploreService) ReloadSubs(dir string) error {
	node, err := f.GetNodeExact(dir)
	if err != nil {
		return err
	}
	return node.reload()
}

// reload dir and return subsCache
func (f *FileExploreService) ReloadSubsCached(dir string, evt ...FesEvent) error {
	node, err := f.GetNodeExact(dir)
	if err != nil {
		return err
	}
	return f.reloadSubsCached(node, evt...)
}

// reload parent's dir and update the cache
func (f *FileExploreService) ReloadPreCached(path string, evt ...FesEvent) error {
	node, err := f.GetNodeExact(path)
	if err != nil {
		return err
	}
	return f.reloadSubsCached(node.Pre(), evt...)
}
func (f *FileExploreService) reloadSubsCached(node *Node, evt ...FesEvent) error {
	if node == nil {
		return fmt.Errorf("node:%w", ErrNilValue)
	}
	if err := node.reload(); err != nil {
		return err
	}
	node.updateCache()

	event := FesUpdateSubs
	if len(evt) > 0 {
		event = evt[0]
	}
	f.app.Event.Emit(event, FesEvtUpdateSubs{
		Path: node.GetPath(),
		Subs: node.SubsCache(),
	})
	return nil
}

// return current [Node.SubsCache] after called [FileExploreService.UpdateView]
func (f *FileExploreService) ReloadRootView() error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.root == nil {
		return ErrNilRoot
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
	f.app.Event.Emit(FesReloadRoot, FesEvtUpdateSubs{
		Path: f.root.GetPath(),
		Subs: f.root.SubsCache(),
	})
	return nil
}

// Delete Node
func (f *FileExploreService) DeleteNode(path string) error {
	node, err := f.GetNodeExact(path)
	if err != nil {
		return err
	}

	path = node.GetPath()
	pre := node.Pre()
	if pre == nil {
		return ErrNodeNilPre
	}
	if err := os.Remove(path); err != nil {
		return err
	}

	pre.DelSub(path)
	f.app.Event.Emit(FesDeletedNode, path)

	return nil
}

// Delete Node
func (f *FileExploreService) DeleteAllNode(path string) error {
	node, err := f.GetNodeExact(path)
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
	f.app.Event.Emit(FesDeletedNode, path)

	return nil
}

// mkdir and reload parent subs
// returns parent node's sub cache
func (f *FileExploreService) AddDirNode(parent string, dirName string) error {
	// parent node
	node, err := f.GetNodeExact(parent)
	if err != nil {
		return err
	}
	if !node.Mode.IsDir() {
		return fmt.Errorf("%s:%w", parent, ErrNodeNotDir)
	}
	if err := fsx.Mkdir(filepath.Join(node.GetPath(), dirName)); err != nil {
		return err
	}
	node.reload()
	node.updateCache()
	f.app.Event.Emit(FesUpdateSubs, FesEvtUpdateSubs{
		Path: node.GetPath(),
		Subs: node.SubsCache(),
	})
	return nil
}

// create file and reload parent subs
// returns parent node's sub cache
func (f *FileExploreService) AddFileNode(parent string, fileName string, mode NodeMode) error {
	node, err := f.GetNodeExact(parent)
	if err != nil {
		return err
	}
	if !node.Mode.IsDir() {
		return fmt.Errorf("%s:%w", parent, ErrNodeNotDir)
	}
	if mode.IsDir() {
		return ErrNodeNotFile
	}
	if err := createFile(filepath.Join(node.GetPath(), fileName), mode); err != nil {
		return err
	}
	node.reload()
	node.updateCache()
	f.app.Event.Emit(FesUpdateSubs, FesEvtUpdateSubs{
		Path: node.GetPath(),
		Subs: node.SubsCache(),
	})
	return nil
}

// create file and set the default file content according to the mode
func createFile(path string, mode NodeMode) error {
	if mode.IsDir() {
		return ErrNodeDir
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0664)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("file already exist:%w", err)
		}
		return err
	}
	defer f.Close()

	var content []byte

	switch m := mode; {
	case m.Has(Request):
		id, err := uid.NewULIDMonotonic()
		if err != nil {
			return err
		}
		req := protocol.Request{ReqMeta: protocol.ReqMeta{ID: id}}
		switch {
		case m.Has(HttpGet):
			req.Method = protocol.MethodGet
		case m.Has(HttpPost):
			req.Method = protocol.MethodPost
		case m.Has(HttpPut):
			req.Method = protocol.MethodPut
		case m.Has(HttpDelete):
			req.Method = protocol.MethodDelete
		case m.Has(HttpHead):
			req.Method = protocol.MethodHead
		case m.Has(HttpPatch):
			req.Method = protocol.MethodPatch
		case m.Has(HttpOptions):
			req.Method = protocol.MethodOptions
		case m.Has(HttpConnect):
			req.Method = protocol.MethodConnect
		case m.Has(HttpTrace):
			req.Method = protocol.MethodTrace
		}
		content, _ = toml.Marshal(req)
	case m.Has(ConfFile):
		content, _ = toml.Marshal(core.Conf{})
	}
	f.Write(content)

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
func (f *FileExploreService) GetNodeExact(path string) (*Node, error) {
	node, err := f.getNode(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	if node.GetPath() != filepath.Clean(path) {
		return nil, ErrNodeNotExist
	}
	return node, nil
}
