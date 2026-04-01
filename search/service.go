package search

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wxk6b1203/file-util-manager/connection"
	"github.com/wxk6b1203/file-util-manager/folder"
)

const EventName = "search:event"

const (
	EventTypeStarted   = "started"
	EventTypeResult    = "result"
	EventTypeError     = "error"
	EventTypeCompleted = "completed"
)

type Request struct {
	Query         string   `json:"query" yaml:"query"`
	ConnectionIDs []string `json:"connectionIds,omitempty" yaml:"connection_ids,omitempty"`
	Root          string   `json:"root,omitempty" yaml:"root,omitempty"`
	MaxResults    int      `json:"maxResults,omitempty" yaml:"max_results,omitempty"`
}

type Result struct {
	ConnectionID   string           `json:"connectionId" yaml:"connection_id"`
	ConnectionName string           `json:"connectionName" yaml:"connection_name"`
	Driver         string           `json:"driver" yaml:"driver"`
	File           *folder.FileInfo `json:"file" yaml:"file"`
}

type Error struct {
	ConnectionID   string `json:"connectionId,omitempty" yaml:"connection_id,omitempty"`
	ConnectionName string `json:"connectionName,omitempty" yaml:"connection_name,omitempty"`
	Message        string `json:"message" yaml:"message"`
}

type Summary struct {
	Query                string `json:"query" yaml:"query"`
	ConnectionCount      int    `json:"connectionCount" yaml:"connection_count"`
	CompletedConnections int    `json:"completedConnections" yaml:"completed_connections"`
	FailedConnections    int    `json:"failedConnections" yaml:"failed_connections"`
	ScannedCount         int    `json:"scannedCount" yaml:"scanned_count"`
	MatchedCount         int    `json:"matchedCount" yaml:"matched_count"`
	ResultLimit          int    `json:"resultLimit,omitempty" yaml:"result_limit,omitempty"`
	LimitReached         bool   `json:"limitReached" yaml:"limit_reached"`
	Cancelled            bool   `json:"cancelled" yaml:"cancelled"`
	DurationMs           int64  `json:"durationMs" yaml:"duration_ms"`
}

type Event struct {
	RequestID string   `json:"requestId" yaml:"request_id"`
	Type      string   `json:"type" yaml:"type"`
	Query     string   `json:"query" yaml:"query"`
	Result    *Result  `json:"result,omitempty" yaml:"result,omitempty"`
	Error     *Error   `json:"error,omitempty" yaml:"error,omitempty"`
	Summary   *Summary `json:"summary,omitempty" yaml:"summary,omitempty"`
}

type activeSearch struct {
	stop context.CancelFunc

	mu        sync.Mutex
	cancelled bool
}

func (a *activeSearch) markCancelled() {
	a.mu.Lock()
	a.cancelled = true
	a.mu.Unlock()
	a.stop()
}

func (a *activeSearch) isCancelled() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.cancelled
}

func (a *activeSearch) stopOnly() {
	a.stop()
}

type Service struct {
	connections      *connection.Service
	configMu         sync.RWMutex
	maxConcurrency   int
	defaultMaxResult int

	mu     sync.Mutex
	active map[string]*activeSearch
}

func NewService(connections *connection.Service, maxConcurrency, defaultMaxResult int) *Service {
	if maxConcurrency <= 0 {
		maxConcurrency = 4
	}
	if defaultMaxResult <= 0 {
		defaultMaxResult = 500
	}

	return &Service{
		connections:      connections,
		maxConcurrency:   maxConcurrency,
		defaultMaxResult: defaultMaxResult,
		active:           make(map[string]*activeSearch),
	}
}

func (s *Service) UpdateDefaults(maxConcurrency, defaultMaxResult int) {
	if maxConcurrency <= 0 {
		maxConcurrency = 4
	}
	if defaultMaxResult <= 0 {
		defaultMaxResult = 500
	}

	s.configMu.Lock()
	s.maxConcurrency = maxConcurrency
	s.defaultMaxResult = defaultMaxResult
	s.configMu.Unlock()
}

func (s *Service) Start(ctx context.Context, req Request, emit func(Event)) (string, error) {
	req.Query = strings.TrimSpace(req.Query)
	if req.Query == "" {
		return "", fmt.Errorf("search query is required")
	}

	defs, err := s.resolveConnections(ctx, req.ConnectionIDs)
	if err != nil {
		return "", err
	}
	if len(defs) == 0 {
		return "", fmt.Errorf("no enabled connections available for search")
	}

	if req.MaxResults <= 0 {
		_, defaultMaxResult := s.defaults()
		req.MaxResults = defaultMaxResult
	}

	requestID := uuid.NewString()
	searchCtx, cancel := context.WithCancel(ctx)
	s.storeActive(requestID, &activeSearch{stop: cancel})

	go s.run(searchCtx, requestID, req, defs, emit)

	return requestID, nil
}

func (s *Service) Cancel(requestID string) error {
	s.mu.Lock()
	active, ok := s.active[requestID]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("search request %q not found", requestID)
	}

	active.markCancelled()
	return nil
}

func (s *Service) run(ctx context.Context, requestID string, req Request, defs []connection.Definition, emit func(Event)) {
	startedAt := time.Now()
	summary := &Summary{
		Query:           req.Query,
		ConnectionCount: len(defs),
		ResultLimit:     req.MaxResults,
	}

	emit(Event{
		RequestID: requestID,
		Type:      EventTypeStarted,
		Query:     req.Query,
		Summary:   cloneSummary(summary),
	})

	var (
		wg          sync.WaitGroup
		maxConc     = s.currentMaxConcurrency()
		semaphore   = make(chan struct{}, maxConc)
		mu          sync.Mutex
		lowerQuery  = strings.ToLower(req.Query)
		searchRoot  = cleanPath(req.Root)
		limit       = req.MaxResults
		limitCancel sync.Once
	)

	for _, def := range defs {
		if ctx.Err() != nil {
			break
		}

		wg.Add(1)
		go func(def connection.Definition) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			case semaphore <- struct{}{}:
			}
			defer func() { <-semaphore }()

			mgr, _, err := s.connections.Manager(ctx, def.ID)
			if err != nil {
				mu.Lock()
				summary.FailedConnections++
				mu.Unlock()
				emit(Event{
					RequestID: requestID,
					Type:      EventTypeError,
					Query:     req.Query,
					Error: &Error{
						ConnectionID:   def.ID,
						ConnectionName: def.Name,
						Message:        err.Error(),
					},
				})
				return
			}

			items, err := mgr.List(ctx, searchRoot, &folder.ListOptions{Recursive: true})
			if err != nil {
				mu.Lock()
				summary.FailedConnections++
				mu.Unlock()
				emit(Event{
					RequestID: requestID,
					Type:      EventTypeError,
					Query:     req.Query,
					Error: &Error{
						ConnectionID:   def.ID,
						ConnectionName: def.Name,
						Message:        err.Error(),
					},
				})
				return
			}

			sort.Slice(items, func(i, j int) bool {
				return strings.ToLower(items[i].Path) < strings.ToLower(items[j].Path)
			})

			for _, item := range items {
				if item == nil || ctx.Err() != nil {
					return
				}

				mu.Lock()
				summary.ScannedCount++
				limitReached := limit > 0 && summary.MatchedCount >= limit
				mu.Unlock()
				if limitReached {
					limitCancel.Do(func() {
						mu.Lock()
						summary.LimitReached = true
						mu.Unlock()
						s.stop(requestID)
					})
					return
				}

				if !matchesQuery(item, lowerQuery) {
					continue
				}

				mu.Lock()
				summary.MatchedCount++
				justReachedLimit := limit > 0 && summary.MatchedCount >= limit
				mu.Unlock()

				emit(Event{
					RequestID: requestID,
					Type:      EventTypeResult,
					Query:     req.Query,
					Result: &Result{
						ConnectionID:   def.ID,
						ConnectionName: def.Name,
						Driver:         def.Driver,
						File:           cloneFileInfo(item),
					},
				})

				if justReachedLimit {
					limitCancel.Do(func() {
						mu.Lock()
						summary.LimitReached = true
						mu.Unlock()
						s.stop(requestID)
					})
					return
				}
			}

			mu.Lock()
			summary.CompletedConnections++
			mu.Unlock()
		}(def)
	}

	wg.Wait()

	mu.Lock()
	summary.DurationMs = time.Since(startedAt).Milliseconds()
	mu.Unlock()

	summary.Cancelled = s.removeActive(requestID)

	emit(Event{
		RequestID: requestID,
		Type:      EventTypeCompleted,
		Query:     req.Query,
		Summary:   cloneSummary(summary),
	})
}

func (s *Service) resolveConnections(ctx context.Context, requested []string) ([]connection.Definition, error) {
	if len(requested) == 0 {
		defs, err := s.connections.List(ctx)
		if err != nil {
			return nil, err
		}

		filtered := make([]connection.Definition, 0, len(defs))
		for _, def := range defs {
			if def.Enabled {
				filtered = append(filtered, def)
			}
		}
		return filtered, nil
	}

	defs := make([]connection.Definition, 0, len(requested))
	for _, id := range requested {
		def, err := s.connections.Get(ctx, strings.TrimSpace(id))
		if err != nil {
			return nil, err
		}
		if def != nil && def.Enabled {
			defs = append(defs, *def)
		}
	}
	return defs, nil
}

func (s *Service) storeActive(requestID string, active *activeSearch) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active[requestID] = active
}

func (s *Service) defaults() (int, int) {
	s.configMu.RLock()
	defer s.configMu.RUnlock()
	return s.maxConcurrency, s.defaultMaxResult
}

func (s *Service) currentMaxConcurrency() int {
	maxConcurrency, _ := s.defaults()
	return maxConcurrency
}

func (s *Service) stop(requestID string) {
	s.mu.Lock()
	active, ok := s.active[requestID]
	s.mu.Unlock()
	if !ok {
		return
	}

	active.stopOnly()
}

func (s *Service) removeActive(requestID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	active, ok := s.active[requestID]
	if !ok {
		return false
	}
	delete(s.active, requestID)
	return active.isCancelled()
}

func matchesQuery(item *folder.FileInfo, lowerQuery string) bool {
	if item == nil {
		return false
	}

	name := strings.ToLower(strings.TrimSpace(item.Name))
	fullPath := strings.ToLower(strings.TrimSpace(item.Path))
	return strings.Contains(name, lowerQuery) || strings.Contains(fullPath, lowerQuery)
}

func cleanPath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "." || trimmed == "/" {
		return ""
	}

	cleaned := path.Clean(strings.ReplaceAll(trimmed, "\\", "/"))
	if cleaned == "." || cleaned == "/" {
		return ""
	}
	return strings.TrimPrefix(cleaned, "/")
}

func cloneSummary(in *Summary) *Summary {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func cloneFileInfo(in *folder.FileInfo) *folder.FileInfo {
	if in == nil {
		return nil
	}
	out := *in
	if in.Owner != nil {
		owner := *in.Owner
		out.Owner = &owner
	}
	if in.Metadata != nil {
		out.Metadata = make(map[string]string, len(in.Metadata))
		for key, value := range in.Metadata {
			out.Metadata[key] = value
		}
	}
	return &out
}
