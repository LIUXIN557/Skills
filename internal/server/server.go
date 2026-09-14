// Package server 提供本地 Web 交互：展示技能列表/启停、触发推送。
// 只读从内存清单取；写操作在文件锁下修改并在锁内保存。
package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strconv"

	"github.com/liuxin/skillhub"
	"github.com/liuxin/skillhub/internal/lock"
	"github.com/liuxin/skillhub/internal/registry"
	"github.com/liuxin/skillhub/internal/sync"
)

// Server 封装根目录与端口。
type Server struct {
	root string
}

// New 构造 Server。
func New(root string) *Server {
	return &Server{root: root}
}

type stateResponse struct {
	Targets      []string          `json:"targets"`
	Sources      []string          `json:"sources"`
	Skills       []*registry.Skill `json:"skills"`
	Patches      map[string]int    `json:"patches"` // skillID -> count
	ConfigStatus string            `json:"configStatus"`
}

// Serve 启动 HTTP 服务并阻塞直到退出。
func (s *Server) Serve(port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/skills/{id}/enable", s.handleSetEnabled(true))
	mux.HandleFunc("/api/skills/{id}/disable", s.handleSetEnabled(false))
	mux.HandleFunc("/api/push", s.handlePush)
	// 静态前端：以嵌入 FS 的 web 子树为根，/ 命中 index.html
	webFS, err := fs.Sub(skillhub.FS, "web")
	if err != nil {
		return err
	}
	staticFS := http.FileServer(http.FS(webFS))
	mux.Handle("/", noDir(staticFS))

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	fmt.Printf("Skill Hub 网页已启动: http://%s\n", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) load() (*registry.File, error) {
	return registry.Load(s.root, false)
}

func (s *Server) writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	reg, err := s.load()
	if err != nil {
		s.writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	counts := map[string]int{}
	for _, p := range reg.Patches {
		counts[p.Skill]++
	}
	resp := stateResponse{
		Targets:      targetPaths(reg),
		Sources:      sourceIDs(reg),
		Skills:       reg.ListSkills(),
		Patches:      counts,
		ConfigStatus: "ok",
	}
	s.writeJSON(w, 200, resp)
}

func (s *Server) handleSetEnabled(value bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		path := registry.ConfigPathFor(s.root)
		var err error
		err = lock.LockedDo(path, func() error {
			reg2, e := registry.Load(s.root, false)
			if e != nil {
				return e
			}
			if e := reg2.SetEnabled(id, value); e != nil {
				return e
			}
			return reg2.Save()
		})
		if err != nil {
			s.writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		s.writeJSON(w, 200, map[string]string{"ok": "true", "id": id})
	}
}

func (s *Server) handlePush(w http.ResponseWriter, r *http.Request) {
	reg, err := s.load()
	if err != nil {
		s.writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	dryRun := r.URL.Query().Get("dryRun") == "1"
	opts := sync.Options{DryRun: dryRun, Verbose: false, Out: os.Stdout}
	_, err = sync.Run(reg, opts)
	if err != nil {
		s.writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, 200, map[string]string{"ok": "true", "dryRun": strconv.FormatBool(dryRun)})
}

func targetPaths(reg *registry.File) []string {
	var out []string
	for _, t := range reg.Targets {
		out = append(out, t.Path)
	}
	return out
}

func sourceIDs(reg *registry.File) []string {
	var out []string
	for _, s := range reg.Sources {
		out = append(out, s.ID)
	}
	return out
}

// noDir 去掉目录列表能力，仅服务文件。
func noDir(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path[len(r.URL.Path)-1:] == "/" {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
