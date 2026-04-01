package connection

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wxk6b1203/file-util-manager/config"
	"github.com/wxk6b1203/file-util-manager/folder"
)

type Definition struct {
	ID          string            `json:"id" yaml:"id"`
	Name        string            `json:"name" yaml:"name"`
	Driver      string            `json:"driver" yaml:"driver"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Enabled     bool              `json:"enabled" yaml:"enabled"`
	ReadOnly    bool              `json:"readOnly,omitempty" yaml:"read_only,omitempty"`
	Root        string            `json:"root,omitempty" yaml:"root,omitempty"`
	Tags        []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Config      map[string]any    `json:"config,omitempty" yaml:"config,omitempty"`
}

type State struct {
	ID              string               `json:"id" yaml:"id"`
	Name            string               `json:"name" yaml:"name"`
	Driver          string               `json:"driver" yaml:"driver"`
	Connected       bool                 `json:"connected" yaml:"connected"`
	LastError       string               `json:"lastError,omitempty" yaml:"last_error,omitempty"`
	LastConnectedAt *time.Time           `json:"lastConnectedAt,omitempty" yaml:"last_connected_at,omitempty"`
	Capabilities    *folder.Capabilities `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
}

type Repository interface {
	List(ctx context.Context) ([]Definition, error)
	Get(ctx context.Context, id string) (*Definition, error)
	Save(ctx context.Context, def Definition) (Definition, error)
	Delete(ctx context.Context, id string) error
}

type FileRepository struct {
	path string
	mu   sync.Mutex
}

type Service struct {
	repo Repository

	mu     sync.RWMutex
	states map[string]State
}

func NewFileRepository(path string) *FileRepository {
	return &FileRepository{path: path}
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:   repo,
		states: make(map[string]State),
	}
}

func (r *FileRepository) List(_ context.Context) ([]Definition, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cfg, err := config.LoadConnectionsConfig(r.path)
	if err != nil {
		return nil, err
	}

	defs := make([]Definition, 0, len(cfg.Config.Connections))
	for _, item := range cfg.Config.Connections {
		defs = append(defs, definitionFromConfig(item))
	}
	sort.Slice(defs, func(i, j int) bool {
		return defs[i].Name < defs[j].Name
	})
	return defs, nil
}

func (r *FileRepository) Get(ctx context.Context, id string) (*Definition, error) {
	defs, err := r.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, def := range defs {
		if def.ID == id {
			copyDef := def
			return &copyDef, nil
		}
	}
	return nil, fmt.Errorf("connection %q: %w", id, folder.ErrNotFound)
}

func (r *FileRepository) Save(_ context.Context, def Definition) (Definition, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	normalized, err := normalizeDefinition(def)
	if err != nil {
		return Definition{}, err
	}

	cfg, err := config.LoadConnectionsConfig(r.path)
	if err != nil {
		return Definition{}, err
	}

	replaced := false
	for i := range cfg.Config.Connections {
		if cfg.Config.Connections[i].ID == normalized.ID {
			cfg.Config.Connections[i] = normalized.toConfig()
			replaced = true
			break
		}
	}
	if !replaced {
		cfg.Config.Connections = append(cfg.Config.Connections, normalized.toConfig())
	}

	if err := config.SaveConnectionsConfig(r.path, cfg.Config); err != nil {
		return Definition{}, err
	}
	return normalized, nil
}

func (r *FileRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cfg, err := config.LoadConnectionsConfig(r.path)
	if err != nil {
		return err
	}

	next := make([]config.ConnectionDefinition, 0, len(cfg.Config.Connections))
	found := false
	for _, item := range cfg.Config.Connections {
		if item.ID == id {
			found = true
			continue
		}
		next = append(next, item)
	}
	if !found {
		return fmt.Errorf("connection %q: %w", id, folder.ErrNotFound)
	}

	cfg.Config.Connections = next
	return config.SaveConnectionsConfig(r.path, cfg.Config)
}

func (s *Service) ListDrivers() []folder.DriverInfo {
	return folder.RegisteredDriverInfo()
}

func (s *Service) List(ctx context.Context) ([]Definition, error) {
	return s.repo.List(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (*Definition, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Save(ctx context.Context, def Definition) (Definition, error) {
	normalized, err := normalizeDefinition(def)
	if err != nil {
		return Definition{}, err
	}
	if _, err := folder.Factory(normalized.Driver); err != nil {
		return Definition{}, err
	}
	return s.repo.Save(ctx, normalized)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.Close(ctx, id); err != nil && !errors.Is(err, folder.ErrNotFound) {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.mu.Lock()
	delete(s.states, id)
	s.mu.Unlock()
	return nil
}

func (s *Service) Open(ctx context.Context, id string) (*State, error) {
	def, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !def.Enabled {
		return nil, fmt.Errorf("connection %q is disabled", id)
	}

	s.mu.RLock()
	current, ok := s.states[id]
	s.mu.RUnlock()
	if ok && current.Connected {
		copyState := current
		return &copyState, nil
	}

	mgr, err := folder.CreateInstance(ctx, def.Driver, def.ID, def.toDriverOptions())
	created := false
	if err != nil {
		if errors.Is(err, folder.ErrAlreadyExist) {
			mgr, err = folder.GetInstance(def.Driver, def.ID)
		}
		if err != nil {
			s.setErrorState(*def, err)
			return nil, err
		}
	} else {
		created = true
	}

	if hc, ok := mgr.(folder.HealthChecker); ok {
		if err := hc.Ping(ctx); err != nil {
			if created {
				_ = folder.DeleteInstance(def.Driver, def.ID)
			}
			s.setErrorState(*def, err)
			return nil, err
		}
	}

	caps := mgr.Capabilities()
	now := time.Now()
	state := State{
		ID:              def.ID,
		Name:            def.Name,
		Driver:          def.Driver,
		Connected:       true,
		LastConnectedAt: &now,
		Capabilities:    &caps,
	}

	s.mu.Lock()
	s.states[id] = state
	s.mu.Unlock()

	return &state, nil
}

func (s *Service) Close(ctx context.Context, id string) error {
	_ = ctx

	s.mu.RLock()
	state, ok := s.states[id]
	s.mu.RUnlock()

	driverName := ""
	if ok && state.Driver != "" {
		driverName = state.Driver
	} else {
		def, err := s.repo.Get(context.Background(), id)
		if err != nil {
			if errors.Is(err, folder.ErrNotFound) {
				return err
			}
			return err
		}
		driverName = def.Driver
	}

	err := folder.DeleteInstance(driverName, id)
	if err != nil && !errors.Is(err, folder.ErrNotFound) {
		s.setCloseError(id, driverName, err)
		return err
	}

	s.mu.Lock()
	prev := s.states[id]
	prev.ID = id
	prev.Driver = driverName
	prev.Connected = false
	prev.Capabilities = nil
	prev.LastError = ""
	s.states[id] = prev
	s.mu.Unlock()
	return nil
}

func (s *Service) ListStates() []State {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]State, 0, len(s.states))
	for _, state := range s.states {
		out = append(out, state)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

func (s *Service) Manager(ctx context.Context, id string) (folder.Manager, *Definition, error) {
	def, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	mgr, err := folder.GetInstance(def.Driver, def.ID)
	if err != nil {
		if errors.Is(err, folder.ErrNotFound) {
			if _, openErr := s.Open(ctx, id); openErr != nil {
				return nil, nil, openErr
			}
			mgr, err = folder.GetInstance(def.Driver, def.ID)
		}
		if err != nil {
			return nil, nil, err
		}
	}

	return mgr, def, nil
}

func (s *Service) setErrorState(def Definition, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prev := s.states[def.ID]
	prev.ID = def.ID
	prev.Name = def.Name
	prev.Driver = def.Driver
	prev.Connected = false
	prev.LastError = err.Error()
	prev.Capabilities = nil
	s.states[def.ID] = prev
}

func (s *Service) setCloseError(id, driverName string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prev := s.states[id]
	prev.ID = id
	prev.Driver = driverName
	prev.Connected = false
	prev.LastError = err.Error()
	prev.Capabilities = nil
	s.states[id] = prev
}

func normalizeDefinition(def Definition) (Definition, error) {
	def.ID = strings.TrimSpace(def.ID)
	if def.ID == "" {
		def.ID = uuid.NewString()
	}
	def.Name = strings.TrimSpace(def.Name)
	def.Driver = strings.TrimSpace(def.Driver)
	def.Description = strings.TrimSpace(def.Description)
	def.Root = strings.TrimSpace(def.Root)

	if def.Name == "" {
		return Definition{}, fmt.Errorf("connection %q: name is required", def.ID)
	}
	if def.Driver == "" {
		return Definition{}, fmt.Errorf("connection %q: driver is required", def.ID)
	}
	if def.Metadata == nil {
		def.Metadata = map[string]string{}
	}
	if def.Config == nil {
		def.Config = map[string]any{}
	}

	return def, nil
}

func definitionFromConfig(def config.ConnectionDefinition) Definition {
	if def.Metadata == nil {
		def.Metadata = map[string]string{}
	}
	if def.Config == nil {
		def.Config = map[string]any{}
	}
	return Definition{
		ID:          def.ID,
		Name:        def.Name,
		Driver:      def.Driver,
		Description: def.Description,
		Enabled:     def.Enabled,
		ReadOnly:    def.ReadOnly,
		Root:        def.Root,
		Tags:        append([]string(nil), def.Tags...),
		Metadata:    cloneStringMap(def.Metadata),
		Config:      cloneAnyMap(def.Config),
	}
}

func (d Definition) toConfig() config.ConnectionDefinition {
	return config.ConnectionDefinition{
		ID:          d.ID,
		Name:        d.Name,
		Driver:      d.Driver,
		Description: d.Description,
		Enabled:     d.Enabled,
		ReadOnly:    d.ReadOnly,
		Root:        d.Root,
		Tags:        append([]string(nil), d.Tags...),
		Metadata:    cloneStringMap(d.Metadata),
		Config:      cloneAnyMap(d.Config),
	}
}

func (d Definition) toDriverOptions() *folder.DriverOptions {
	return &folder.DriverOptions{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		Driver:      d.Driver,
		Root:        d.Root,
		Enabled:     d.Enabled,
		ReadOnly:    d.ReadOnly,
		Tags:        append([]string(nil), d.Tags...),
		Metadata:    cloneStringMap(d.Metadata),
		Config:      cloneAnyMap(d.Config),
	}
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func cloneAnyMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
