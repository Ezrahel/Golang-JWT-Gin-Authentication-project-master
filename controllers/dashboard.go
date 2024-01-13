package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/akhil/golang-jwt-project/models"
	"github.com/gin-gonic/gin"
)

type Appointment struct {
	Duration time.Duration `json:"duration"`
	models.Doctor
	DayOfWeek time.Weekday `json:"dayOfWeek"`
	Purpose   string       `json:"purpose"`
}

type DashboardController struct{}

func (u *models.Doctor) GetUserRole() string {
	if u.Doctor_type == nil {
		return "Unknown"
	}

	ctx := gin.Default()
	ctx.LoadHTMLGlob("frontend/*html")
	userType := strings.ToUpper(*u.Doctor_type)
	switch userType {
	case "ADMIN":
		ctx.GET("/dashboard", func(c *gin.Context) {
			c.HTML(http.StatusOK, "staffdashboard", gin.H{
				"Message": "Welcome to the Staff dashboard",
			})
		})
	case "PATIENT":
		ctx.GET("/patient-dashboard", func(c *gin.Context) {
			c.HTML(http.StatusOK, "dashboard", gin.H{
				"Message": "Welcome to the patient dashboard",
			})
		})
	case "NURSE":
		ctx.GET("/dashboard", func(c *gin.Context) {
			c.HTML(http.StatusOK, "staffdashboard", gin.H{
				"Message": "Welcome to the Staff dashboard",
			})
		})
	default:
		ctx.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "landingpage", gin.H{
				"Message": "Welcome to the Landing Page",
			})
		})
	}
}

//Pending functionalities
//Logout
//Search
//Medical Records(Surgeries and others)
//Appointment
//Hospital room no
