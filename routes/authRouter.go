package routes

import (
	controller "github.com/akhil/golang-jwt-project/controllers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("doctors/signup", controller.Signup())
	incomingRoutes.POST("doctors/login", controller.Login())
}
