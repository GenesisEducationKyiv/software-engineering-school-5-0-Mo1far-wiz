package internal_test

import (
	"testing"

	"github.com/matthewmcnew/archtest"
)

// Domain/Models layer tests - Models should be independent of everything
func Test_Models_ShouldNotDependOn_API(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/models").ShouldNotDependOn(
		"weather-subscription/internal/api",
	)
}

func Test_Models_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/models").ShouldNotDependOn(
		"weather-subscription/internal/api/handlers",
	)
}

func Test_Models_ShouldNotDependOn_Store(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/models").ShouldNotDependOn(
		"weather-subscription/internal/store",
	)
}

func Test_Models_ShouldNotDependOn_Weather(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/models").ShouldNotDependOn(
		"weather-subscription/internal/weather",
	)
}

func Test_Models_ShouldNotDependOn_Mailer(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/models").ShouldNotDependOn(
		"weather-subscription/internal/mailer",
	)
}

func Test_Models_ShouldNotDependOn_Cache(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/models").ShouldNotDependOn(
		"weather-subscription/internal/cache",
	)
}

func Test_Models_ShouldNotDependOn_Database(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/models").ShouldNotDependOn(
		"weather-subscription/internal/database",
	)
}

func Test_Models_ShouldNotDependOn_Redis(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/models").ShouldNotDependOn(
		"weather-subscription/internal/redis",
	)
}

func Test_Models_ShouldNotDependOn_Application(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/models").ShouldNotDependOn(
		"weather-subscription/internal/application",
	)
}

// Store layer tests - Store should not depend on higher layers
func Test_Store_ShouldNotDependOn_API(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/store").ShouldNotDependOn(
		"weather-subscription/internal/api",
	)
}

func Test_Store_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/store").ShouldNotDependOn(
		"weather-subscription/internal/api/handlers",
	)
}

func Test_Store_ShouldNotDependOn_Weather(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/store").ShouldNotDependOn(
		"weather-subscription/internal/weather",
	)
}

func Test_Store_ShouldNotDependOn_Mailer(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/store").ShouldNotDependOn(
		"weather-subscription/internal/mailer",
	)
}

func Test_Store_ShouldNotDependOn_Cache(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/store").ShouldNotDependOn(
		"weather-subscription/internal/cache",
	)
}

func Test_Store_ShouldNotDependOn_Application(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/store").ShouldNotDependOn(
		"weather-subscription/internal/application",
	)
}

// Weather service tests - Weather should not depend on higher layers
func Test_Weather_ShouldNotDependOn_API(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/weather").ShouldNotDependOn(
		"weather-subscription/internal/api",
	)
}

func Test_Weather_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/weather").ShouldNotDependOn(
		"weather-subscription/internal/api/handlers",
	)
}

func Test_Weather_ShouldNotDependOn_Store(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/weather").ShouldNotDependOn(
		"weather-subscription/internal/store",
	)
}

func Test_Weather_ShouldNotDependOn_Mailer(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/weather").ShouldNotDependOn(
		"weather-subscription/internal/mailer",
	)
}

func Test_Weather_ShouldNotDependOn_Application(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/weather").ShouldNotDependOn(
		"weather-subscription/internal/application",
	)
}

// Cache service tests - Cache should not depend on higher layers
func Test_Cache_ShouldNotDependOn_API(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/cache").ShouldNotDependOn(
		"weather-subscription/internal/api",
	)
}

func Test_Cache_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/cache").ShouldNotDependOn(
		"weather-subscription/internal/api/handlers",
	)
}

func Test_Cache_ShouldNotDependOn_Store(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/cache").ShouldNotDependOn(
		"weather-subscription/internal/store",
	)
}

func Test_Cache_ShouldNotDependOn_Weather(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/cache").ShouldNotDependOn(
		"weather-subscription/internal/weather",
	)
}

func Test_Cache_ShouldNotDependOn_Mailer(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/cache").ShouldNotDependOn(
		"weather-subscription/internal/mailer",
	)
}

func Test_Cache_ShouldNotDependOn_Application(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/cache").ShouldNotDependOn(
		"weather-subscription/internal/application",
	)
}

// Mailer service tests - Mailer should not depend on higher layers
func Test_Mailer_ShouldNotDependOn_API(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/mailer").ShouldNotDependOn(
		"weather-subscription/internal/api",
	)
}

func Test_Mailer_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/mailer").ShouldNotDependOn(
		"weather-subscription/internal/api/handlers",
	)
}

func Test_Mailer_ShouldNotDependOn_Store(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/mailer").ShouldNotDependOn(
		"weather-subscription/internal/store",
	)
}

func Test_Mailer_ShouldNotDependOn_Application(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/mailer").ShouldNotDependOn(
		"weather-subscription/internal/application",
	)
}

// Infrastructure layer tests - Infrastructure should not depend on higher layers
func Test_Database_ShouldNotDependOn_API(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/database").ShouldNotDependOn(
		"weather-subscription/internal/api",
	)
}

func Test_Database_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/database").ShouldNotDependOn(
		"weather-subscription/internal/api/handlers",
	)
}

func Test_Database_ShouldNotDependOn_Store(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/database").ShouldNotDependOn(
		"weather-subscription/internal/store",
	)
}

func Test_Database_ShouldNotDependOn_Weather(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/database").ShouldNotDependOn(
		"weather-subscription/internal/weather",
	)
}

func Test_Database_ShouldNotDependOn_Mailer(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/database").ShouldNotDependOn(
		"weather-subscription/internal/mailer",
	)
}

func Test_Database_ShouldNotDependOn_Cache(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/database").ShouldNotDependOn(
		"weather-subscription/internal/cache",
	)
}

func Test_Database_ShouldNotDependOn_Application(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/database").ShouldNotDependOn(
		"weather-subscription/internal/application",
	)
}

// Configuration tests - Config should be independent
func Test_Config_ShouldNotDependOn_API(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/config").ShouldNotDependOn(
		"weather-subscription/internal/api",
	)
}

func Test_Config_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/config").ShouldNotDependOn(
		"weather-subscription/internal/api/handlers",
	)
}

func Test_Config_ShouldNotDependOn_Store(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/config").ShouldNotDependOn(
		"weather-subscription/internal/store",
	)
}

func Test_Config_ShouldNotDependOn_Weather(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/config").ShouldNotDependOn(
		"weather-subscription/internal/weather",
	)
}

func Test_Config_ShouldNotDependOn_Mailer(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/config").ShouldNotDependOn(
		"weather-subscription/internal/mailer",
	)
}

func Test_Config_ShouldNotDependOn_Cache(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/config").ShouldNotDependOn(
		"weather-subscription/internal/cache",
	)
}

func Test_Config_ShouldNotDependOn_Database(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/config").ShouldNotDependOn(
		"weather-subscription/internal/database",
	)
}

func Test_Config_ShouldNotDependOn_Redis(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/config").ShouldNotDependOn(
		"weather-subscription/internal/redis",
	)
}

func Test_Config_ShouldNotDependOn_Application(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/config").ShouldNotDependOn(
		"weather-subscription/internal/application",
	)
}

// Error handling tests - Errors should be independent
func Test_Svc_ShouldNotDependOn_API(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/svc").ShouldNotDependOn(
		"weather-subscription/internal/api",
	)
}

func Test_Svc_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/svc").ShouldNotDependOn(
		"weather-subscription/internal/api/handlers",
	)
}

func Test_Svc_ShouldNotDependOn_Store(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/svc").ShouldNotDependOn(
		"weather-subscription/internal/store",
	)
}

func Test_Svc_ShouldNotDependOn_Weather(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/svc").ShouldNotDependOn(
		"weather-subscription/internal/weather",
	)
}

func Test_Svc_ShouldNotDependOn_Mailer(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/svc").ShouldNotDependOn(
		"weather-subscription/internal/mailer",
	)
}

func Test_Svc_ShouldNotDependOn_Cache(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/svc").ShouldNotDependOn(
		"weather-subscription/internal/cache",
	)
}

func Test_Svc_ShouldNotDependOn_Database(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/svc").ShouldNotDependOn(
		"weather-subscription/internal/database",
	)
}

func Test_Svc_ShouldNotDependOn_Redis(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/svc").ShouldNotDependOn(
		"weather-subscription/internal/redis",
	)
}

func Test_Svc_ShouldNotDependOn_Application(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/svc").ShouldNotDependOn(
		"weather-subscription/internal/application",
	)
}

// API layer tests - API should not depend on specific implementations
func Test_API_ShouldNotDependOn_Database(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/api").ShouldNotDependOn(
		"weather-subscription/internal/database",
	)
}

func Test_API_ShouldNotDependOn_Redis(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/api").ShouldNotDependOn(
		"weather-subscription/internal/redis",
	)
}

func Test_Handlers_ShouldNotDependOn_Database(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/api/handlers").ShouldNotDependOn(
		"weather-subscription/internal/database",
	)
}

func Test_Handlers_ShouldNotDependOn_Redis(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/api/handlers").ShouldNotDependOn(
		"weather-subscription/internal/redis",
	)
}

func Test_Handlers_ShouldNotDependOn_Cache(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/api/handlers").ShouldNotDependDirectlyOn(
		"weather-subscription/internal/cache",
	)
}

func Test_Handlers_ShouldNotDependOn_Store(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/api/handlers").ShouldNotDependOn(
		"weather-subscription/internal/store",
	)
}

func Test_Handlers_ShouldNotDependOn_Application(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/api/handlers").ShouldNotDependOn(
		"weather-subscription/internal/application",
	)
}

// Middleware tests - Middleware should be independent
func Test_Middleware_ShouldNotDependOn_Handlers(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/api/middleware").ShouldNotDependOn(
		"weather-subscription/internal/api/handlers",
	)
}

func Test_Middleware_ShouldNotDependOn_Store(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/api/middleware").ShouldNotDependOn(
		"weather-subscription/internal/store",
	)
}

func Test_Middleware_ShouldNotDependOn_Weather(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/api/middleware").ShouldNotDependOn(
		"weather-subscription/internal/weather",
	)
}

func Test_Middleware_ShouldNotDependOn_Mailer(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/api/middleware").ShouldNotDependOn(
		"weather-subscription/internal/mailer",
	)
}

func Test_Middleware_ShouldNotDependOn_Cache(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/api/middleware").ShouldNotDependOn(
		"weather-subscription/internal/cache",
	)
}

func Test_Middleware_ShouldNotDependOn_Database(t *testing.T) {
	archtest.Package(t, "weather-subscription/internal/api/middleware").ShouldNotDependOn(
		"weather-subscription/internal/database",
	)
}

func Test_Middleware_ShouldNotDependOn_Redis(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/api/middleware").ShouldNotDependOn(
		"weather-subscription/internal/redis",
	)
}

func Test_Middleware_ShouldNotDependOn_Application(t *testing.T) {
	t.Parallel()
	archtest.Package(t, "weather-subscription/internal/api/middleware").ShouldNotDependOn(
		"weather-subscription/internal/application",
	)
}
