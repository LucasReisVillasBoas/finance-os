package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// planLevels maps plan names to numeric levels for comparison.
var planLevels = map[string]int{
	"free":    0,
	"pro":     1,
	"premium": 2,
}

// PlanMiddleware returns a Gin middleware that enforces a minimum plan level.
func PlanMiddleware(minPlan string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userPlan := c.GetString("user_plan")
		if userPlan == "" {
			userPlan = "free"
		}

		required := planLevels[minPlan]
		actual := planLevels[userPlan]

		if actual < required {
			c.AbortWithStatusJSON(http.StatusPaymentRequired, gin.H{
				"error": gin.H{
					"code":          "PLAN_REQUIRED",
					"message":       fmt.Sprintf("Esta funcionalidade requer o plano %s ou superior", minPlan),
					"required_plan": minPlan,
				},
			})
			return
		}
		c.Next()
	}
}
