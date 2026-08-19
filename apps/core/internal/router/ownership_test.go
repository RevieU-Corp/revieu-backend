package router

import (
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/config"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/admin"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/ai"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/auth"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/category"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/conversation"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/coupon"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/feed"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/media"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/merchant"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/merchants"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/notification"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/order"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/payment"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/review"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/stores"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/user"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/users"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/verification"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/voucher"
)

type registrar struct {
	name string
	fn   func(*gin.RouterGroup, *config.Config)
}

// allDomains mirrors Setup's registration list. Keeping it here lets the test
// attribute each route to the domain that registered it.
func allDomains() []registrar {
	return []registrar{
		{"auth", auth.RegisterRoutes},
		{"ai", ai.RegisterRoutes},
		{"user", user.RegisterRoutes},
		{"users", users.RegisterRoutes},
		{"coupon", coupon.RegisterRoutes},
		{"feed", feed.RegisterRoutes},
		{"merchant", merchant.RegisterRoutes},
		{"merchants", merchants.RegisterRoutes},
		{"media", media.RegisterRoutes},
		{"payment", payment.RegisterRoutes},
		{"review", review.RegisterRoutes},
		{"voucher", voucher.RegisterRoutes},
		{"stores", stores.RegisterRoutes},
		{"category", category.RegisterRoutes},
		{"conversation", conversation.RegisterRoutes},
		{"notification", notification.RegisterRoutes},
		{"verification", verification.RegisterRoutes},
		{"admin", admin.RegisterRoutes},
		{"order", order.RegisterRoutes},
	}
}

// ownershipOf attributes every route to the domains that register it. Each
// domain registers into its own engine, because gin.Routes() does not preserve
// registration order and therefore cannot be sliced to attribute routes.
func ownershipOf(t *testing.T) map[string][]string {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cfg := contractCfg()
	owners := map[string][]string{}
	for _, d := range allDomains() {
		r := gin.New()
		api := r.Group(cfg.Server.APIBasePath)
		d.fn(api, cfg)
		for _, ri := range r.Routes() {
			key := ri.Method + " " + ri.Path
			owners[key] = append(owners[key], d.name)
		}
	}
	return owners
}

// resourceOf returns the top-level resource segment a route belongs to,
// treating /merchant/<sub> as its own resource since merchant-scoped
// management is a distinct surface from public reads.
func resourceOf(path, basePath string) string {
	rest := strings.TrimPrefix(path, basePath+"/")
	segs := strings.Split(rest, "/")
	if len(segs) == 0 {
		return ""
	}
	if segs[0] == "merchant" && len(segs) > 1 {
		return "merchant/" + segs[1]
	}
	return segs[0]
}

// TestEachResourceHasSingleOwner is the ownership contract: a top-level
// resource prefix must be registered by exactly one domain. Without this,
// "where do I add a route for X" has no answer and route conflicts appear
// as registration-order bugs.
func TestEachResourceHasSingleOwner(t *testing.T) {
	routeOwners := ownershipOf(t)

	owners := map[string]map[string]bool{}
	for route, doms := range routeOwners {
		parts := strings.SplitN(route, " ", 2)
		res := resourceOf(parts[1], "/api/v1")
		if res == "health" || res == "swagger" {
			continue
		}
		if owners[res] == nil {
			owners[res] = map[string]bool{}
		}
		for _, d := range doms {
			owners[res][d] = true
		}
	}

	resources := make([]string, 0, len(owners))
	for res := range owners {
		resources = append(resources, res)
	}
	sort.Strings(resources)

	for _, res := range resources {
		doms := owners[res]
		if len(doms) > 1 {
			names := make([]string, 0, len(doms))
			for d := range doms {
				names = append(names, d)
			}
			sort.Strings(names)
			t.Errorf("resource /%s is registered by %d domains: %s (must have a single owner)",
				res, len(names), strings.Join(names, ", "))
		}
	}
}
