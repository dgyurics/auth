package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	r := require.New(t)

	c := New()

	r.Equal("*", c.Cors.AllowOrigin, "Default CORS allow origin not set correctly")
	r.Equal("GET, POST, OPTIONS", c.Cors.AllowMethods, "Default CORS allow methods not set correctly")
	r.Equal("*", c.Cors.AllowHeaders, "Default CORS allow headers not set correctly")
	r.Equal("true", c.Cors.AllowCredentials, "Default CORS allow credentials not set correctly")

	r.Equal("auth", c.PostgreSQL.Dbname, "Default PostgreSQL dbname not set correctly")
	r.Equal("postgres", c.PostgreSQL.User, "Default PostgreSQL user not set correctly")
	r.Equal("postgres", c.PostgreSQL.Password, "Default PostgreSQL password not set correctly")
	r.Equal("localhost", c.PostgreSQL.Host, "Default PostgreSQL host not set correctly")
	r.Equal(5432, c.PostgreSQL.Port, "Default PostgreSQL port not set correctly")
	r.Equal("disable", c.PostgreSQL.Sslmode, "Default PostgreSQL sslmode not set correctly")
	r.Equal("golang_auth_service", c.PostgreSQL.FallbackApplication, "Default PostgreSQL fallback application not set correctly")

	r.Equal("localhost:6379", c.Redis.Addr, "Default Redis address not set correctly")
	r.Equal("", c.Redis.Username, "Default Redis username not set correctly")
	r.Equal("", c.Redis.Password, "Default Redis password not set correctly")
	r.Equal(0, c.Redis.DB, "Default Redis DB not set correctly")

	r.Equal(30, int(c.RequestTimeout), "Default request timeout not set correctly")

	r.Equal("8080", c.ServerConfig.Port, "Default server port not set correctly")

	r.Equal("X-Session-ID", c.Session.Name, "Default session name not set correctly")
	r.Equal("localhost", c.Session.Domain, "Default session domain not set correctly")
	r.Equal("/", c.Session.Path, "Default session path not set correctly")
	r.False(c.Session.Secure, "Default session secure flag not set correctly")
	r.True(c.Session.HTTPOnly, "Default session HTTPOnly flag not set correctly")
	r.Equal("Strict", c.Session.SameSite, "Default session SameSite not set correctly")
	r.Equal(86400, c.Session.MaxAge, "Default session max age not set correctly")
}
