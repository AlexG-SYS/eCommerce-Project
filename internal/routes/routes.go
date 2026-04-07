package routes

import (
	"log/slog"
	"net/http"

	"database/sql"

	"github.com/AlexG-SYS/eCommerce-Project/internal/data"
	"github.com/AlexG-SYS/eCommerce-Project/internal/handlers"
	"github.com/AlexG-SYS/eCommerce-Project/internal/helpers"
	"github.com/AlexG-SYS/eCommerce-Project/internal/mailer"
	"github.com/AlexG-SYS/eCommerce-Project/internal/middleware"
)

func SetupRoutes(db *sql.DB, logger *slog.Logger, mailer mailer.Mailer, rps float64, burst int, enabled bool, origins []string) http.Handler {
	models := data.NewModels(db)

	app := &helpers.Application{Logger: logger, Mailer: mailer, Models: models}

	h := &handlers.Handler{
		App:    app,
		Models: models,
	}

	mw := middleware.Middleware{
		App:            app,
		LimiterRPS:     rps,
		LimiterBurst:   burst,
		LimiterEnabled: enabled,
		TrustedOrigins: origins,
	}

	mux := http.NewServeMux()

	// --- PUBLIC ROUTES (No Token Required) ---
	mux.HandleFunc("POST /v1/users/login", h.LoginHandler)
	mux.HandleFunc("POST /v1/profiles", h.CreateProfileHandler) // Registration
	mux.HandleFunc("GET /v1/users/activated", h.ActivateUserHandler)
	mux.HandleFunc("GET /v1/products", h.ListProductsHandler)
	mux.HandleFunc("GET /v1/products/{id}", h.GetProductHandler)
	mux.HandleFunc("GET /v1/categories", h.ListCategoriesHandler)

	// --- AUTHENTICATED ROUTES (Customer or Admin) ---
	// Using "Customer" role here as a baseline; Admins bypass this check in your middleware.
	mux.HandleFunc("GET /v1/profiles/me", mw.RequireRole("Customer", h.GetMyProfile))
	mux.HandleFunc("POST /v1/orders", mw.RequireRole("Customer", h.CreateOrderHandler))
	mux.HandleFunc("GET /v1/orders/{id}", mw.RequireRole("Customer", h.GetOrderHandler))

	// --- ADMIN ONLY ROUTES ---
	// Inventory and Catalog Management
	mux.HandleFunc("POST /v1/categories", mw.RequireRole("Admin", h.CreateCategoryHandler))
	mux.HandleFunc("PATCH /v1/categories/{id}", mw.RequireRole("Admin", h.UpdateCategoryHandler))

	mux.HandleFunc("POST /v1/locations", mw.RequireRole("Admin", h.CreateLocationHandler))
	mux.HandleFunc("GET /v1/locations", mw.RequireRole("Admin", h.ListLocationsHandler))
	mux.HandleFunc("PATCH /v1/locations/{id}", mw.RequireRole("Admin", h.UpdateLocationHandler))

	mux.HandleFunc("POST /v1/products", mw.RequireRole("Admin", h.CreateProductHandler))
	mux.HandleFunc("PATCH /v1/products/{id}", mw.RequireRole("Admin", h.UpdateProductHandler))

	mux.HandleFunc("POST /v1/variants", mw.RequireRole("Admin", h.CreateVariantHandler))
	mux.HandleFunc("PATCH /v1/variants/{id}", mw.RequireRole("Admin", h.UpdateVariantHandler))

	mux.HandleFunc("POST /v1/inventory", mw.RequireRole("Admin", h.CreateInventoryHandler))
	mux.HandleFunc("PATCH /v1/inventory/{id}", mw.RequireRole("Admin", h.UpdateInventoryHandler))

	mux.HandleFunc("GET /v1/metrics", mw.RequireRole("Admin", h.MetricsHandler))

	// --- MIXED PERMISSIONS / HELPERS ---
	// Profiles by ID and Shipping might be Admin only or restricted to the owner
	mux.HandleFunc("GET /v1/profiles/{id}", mw.RequireRole("Customer", h.GetProfileHandler))
	mux.HandleFunc("PATCH /v1/profiles/{id}", mw.RequireRole("Customer", h.UpdateProfileHandler))

	mux.HandleFunc("POST /v1/shipping", mw.RequireRole("Admin", h.CreateShippingHandler))
	mux.HandleFunc("PATCH /v1/shipping/{id}", mw.RequireRole("Admin", h.UpdateShippingHandler))
	mux.HandleFunc("GET /v1/shipping/{id}", mw.RequireRole("Customer", h.GetShippingHandler))

	// --- FUTURE ENDPOINTS ---

	// Middleware Chain (Executed bottom to top)
	return mw.Metrics(
		mw.RateLimit( // 1. First, check if they are allowed in
			mw.Logger( // 2. Then, log the request details
				mw.Compress( // 3. Then, prepare to compress the response
					mw.EnableCORS(
						mw.Authenticate(mux),
					), // 4. Finally, handle CORS and Routing
				),
			),
		),
	)
}
