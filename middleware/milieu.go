package middleware

import (
	"git.jagtech.io/Impala/corelib"
	"github.com/gin-gonic/gin"
)

// SetupMilieu provides an easy interface for injecting Milieu into the core handler functions of gin
func SetupMilieu(milieu corelib.Milieu) gin.HandlerFunc {
	return func(c *gin.Context) {
		m := milieu.Clone()
		c.Set("MILIEU", m)
		c.Next()
		m.Cleanup()
	}
}

// MustGetMilieu wraps the default gin context but automatically wraps the output to the correct type for ease
func MustGetMilieu(c *gin.Context) corelib.Milieu {
	return c.MustGet("MILIEU").(corelib.Milieu)
}
