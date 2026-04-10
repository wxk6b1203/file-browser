package webdavtest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	pathpkg "path"
	"testing"

	xwebdav "golang.org/x/net/webdav"
)

type Server struct {
	*httptest.Server
	FS xwebdav.FileSystem
}

func NewServer(t *testing.T, username, password, bearerToken string) *Server {
	t.Helper()

	ctx := context.Background()
	fs := xwebdav.NewMemFS()
	lock := xwebdav.NewMemLS()
	handler := &xwebdav.Handler{
		FileSystem: fs,
		LockSystem: lock,
	}

	authWrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case bearerToken != "":
			if got := r.Header.Get("Authorization"); got != "Bearer "+bearerToken {
				w.Header().Set("WWW-Authenticate", "Bearer")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		case username != "":
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.Header().Set("WWW-Authenticate", `Basic realm="test"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}
		handler.ServeHTTP(w, r)
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = pathpkg.Clean(r.URL.Path)
		authWrapped.ServeHTTP(w, r)
	}))

	if err := fs.Mkdir(ctx, "/scoped", 0o755); err != nil && !os.IsExist(err) {
		server.Close()
		t.Fatalf("mkdir scoped: %v", err)
	}

	return &Server{
		Server: server,
		FS:     fs,
	}
}
