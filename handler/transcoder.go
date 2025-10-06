package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	DB  *gorm.DB
	CFG *gorm.Config
}

func (a *App) registerRoutes(r *gin.Engine) {
	r.LoadHTMLGlob("/templates/*")
}