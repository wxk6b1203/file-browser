package folder

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// DriverOptions contains cross-driver common settings plus backend-specific config.
type DriverOptions struct {
	ID          string            `json:"id,omitempty" yaml:"id,omitempty"`
	Name        string            `json:"name,omitempty" yaml:"name,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Driver      string            `json:"driver,omitempty" yaml:"driver,omitempty"`
	Root        string            `json:"root,omitempty" yaml:"root,omitempty"`
	Enabled     bool              `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	ReadOnly    bool              `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	TimeoutSec  int               `json:"timeoutSec,omitempty" yaml:"timeoutSec,omitempty"`
	Tags        []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Config      map[string]any    `json:"config,omitempty" yaml:"config,omitempty"`
}

// normalize fills in defaults and returns a shallow copy.
func (o *DriverOptions) normalize(driverName, instanceName string) *DriverOptions {
	if o == nil {
		o = &DriverOptions{}
	}
	out := *o
	if out.Driver == "" {
		out.Driver = driverName
	}
	if out.Name == "" {
		out.Name = instanceName
	}
	if out.Config == nil {
		out.Config = map[string]any{}
	}
	return &out
}

// DriverFactory creates a Manager from common options.
type DriverFactory func(ctx context.Context, opt *DriverOptions) (Manager, error)

type driverEntry struct {
	factory   DriverFactory
	instances map[string]Manager
}

var (
	registryMu sync.RWMutex
	registry   = make(map[string]*driverEntry)
)

// ---------------------------------------------------------------------------
// Driver type registration
// ---------------------------------------------------------------------------

// Register registers a driver factory under the given name.
func Register(name string, factory DriverFactory) error {
	if name == "" {
		return fmt.Errorf("register driver: name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("register driver %q: factory cannot be nil", name)
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	if _, ok := registry[name]; ok {
		return fmt.Errorf("register driver %q: already registered", name)
	}
	registry[name] = &driverEntry{factory: factory, instances: make(map[string]Manager)}
	return nil
}

// MustRegister is like Register but panics on error.
func MustRegister(name string, factory DriverFactory) {
	if err := Register(name, factory); err != nil {
		panic(err)
	}
}

// RegisterDriver is a generic helper that eliminates the repetitive
// DecodeConfig+MustRegister boilerplate in every driver's init().
//
// Usage (in driver init):
//
//	folder.RegisterDriver[sftp.Options]("sftp", sftp.New)
func RegisterDriver[T any](name string, newFn func(context.Context, *DriverOptions, *T) (Manager, error)) {
	MustRegister(name, func(ctx context.Context, opt *DriverOptions) (Manager, error) {
		cfg := new(T)
		if opt != nil && opt.Config != nil {
			if err := DecodeConfig(opt.Config, cfg); err != nil {
				return nil, err
			}
		}
		return newFn(ctx, opt, cfg)
	})
}

// Factory returns a registered driver factory.
func Factory(name string) (DriverFactory, error) {
	registryMu.RLock()
	entry, ok := registry[name]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupported, name)
	}
	return entry.factory, nil
}

// RegisteredDrivers returns all registered driver names in sorted order.
func RegisteredDrivers() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ---------------------------------------------------------------------------
// Instance lifecycle
// ---------------------------------------------------------------------------

// NewManager creates a one-off Manager without registering it as an instance.
func NewManager(ctx context.Context, driverName string, opt *DriverOptions) (Manager, error) {
	factory, err := Factory(driverName)
	if err != nil {
		return nil, err
	}
	return factory(ctx, opt.normalize(driverName, ""))
}

// CreateInstance creates a named instance and registers it in the registry.
func CreateInstance(ctx context.Context, driverName, instanceName string, opt *DriverOptions) (Manager, error) {
	if instanceName == "" {
		return nil, fmt.Errorf("create instance: name cannot be empty")
	}

	factory, err := Factory(driverName)
	if err != nil {
		return nil, err
	}

	mgr, err := factory(ctx, opt.normalize(driverName, instanceName))
	if err != nil {
		return nil, err
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	entry, ok := registry[driverName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupported, driverName)
	}
	if _, exists := entry.instances[instanceName]; exists {
		// Close the newly created manager to avoid resource leak.
		if c, ok := mgr.(Closer); ok {
			_ = c.Close()
		}
		return nil, fmt.Errorf("%w: %s/%s", ErrAlreadyExist, driverName, instanceName)
	}
	entry.instances[instanceName] = mgr
	return mgr, nil
}

// GetInstance retrieves a previously registered instance.
func GetInstance(driverName, instanceName string) (Manager, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	entry, ok := registry[driverName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupported, driverName)
	}
	mgr, ok := entry.instances[instanceName]
	if !ok {
		return nil, fmt.Errorf("%w: %s/%s", ErrNotFound, driverName, instanceName)
	}
	return mgr, nil
}

// DeleteInstance removes a registered instance and closes it if it implements Closer.
func DeleteInstance(driverName, instanceName string) error {
	registryMu.Lock()
	defer registryMu.Unlock()

	entry, ok := registry[driverName]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnsupported, driverName)
	}
	mgr, exists := entry.instances[instanceName]
	if !exists {
		return fmt.Errorf("%w: %s/%s", ErrNotFound, driverName, instanceName)
	}
	delete(entry.instances, instanceName)

	// Release resources held by the driver (e.g. SFTP connection).
	if c, ok := mgr.(Closer); ok {
		return c.Close()
	}
	return nil
}

// ListInstances returns all instance names for a driver type, sorted.
func ListInstances(driverName string) ([]string, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	entry, ok := registry[driverName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupported, driverName)
	}

	names := make([]string, 0, len(entry.instances))
	for name := range entry.instances {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}
