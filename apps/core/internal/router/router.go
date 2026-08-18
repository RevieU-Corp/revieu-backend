package router

import (
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/config"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/admin"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/ai"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/auth"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/category"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/content"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/conversation"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/coupon"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/feed"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/follow"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/media"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/merchant"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/notification"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/order"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/payment"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/profile"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/review"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/store"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/user"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/verification"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/voucher"
	"github.com/gin-gonic/gin"
)

// Setup registers all domain routes under the API base path.
func Setup(router *gin.Engine, cfg *config.Config) {
	api := router.Group(cfg.Server.APIBasePath)

	auth.RegisterRoutes(api, cfg)
	ai.RegisterRoutes(api, cfg)
	user.RegisterRoutes(api, cfg)
	profile.RegisterRoutes(api, cfg)
	follow.RegisterRoutes(api, cfg)
	content.RegisterRoutes(api, cfg)
	coupon.RegisterRoutes(api, cfg)
	feed.RegisterRoutes(api, cfg)
	merchant.RegisterRoutes(api, cfg)
	media.RegisterRoutes(api, cfg)
	payment.RegisterRoutes(api, cfg)
	review.RegisterRoutes(api, cfg)
	voucher.RegisterRoutes(api, cfg)
	store.RegisterRoutes(api, cfg)
	category.RegisterRoutes(api, cfg)
	conversation.RegisterRoutes(api, cfg)
	notification.RegisterRoutes(api, cfg)
	verification.RegisterRoutes(api, cfg)
	admin.RegisterRoutes(api, cfg)
	order.RegisterRoutes(api, cfg)
}
