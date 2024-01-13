package routes

import (
	controller "github.com/akhil/golang-jwt-project/controllers"
	"github.com/akhil/golang-jwt-project/middleware"
	"github.com/gin-gonic/gin"
)

func DoctorRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/doctors", controller.GetDoctors())
	incomingRoutes.GET("/doctors/:doctor_id", controller.GetDoctor())
	incomingRoutes.GET("/doctors/:doctor_dashboard", controller.Index)

}
