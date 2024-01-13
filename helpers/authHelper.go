package helper

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func CheckDoctorType(c *gin.Context, role string) (err error) {
	DoctorType := c.GetString("doctor_type")
	err = nil
	if DoctorType != role {
		err = errors.New("Unauthorized to access this resource")
		return err
	}
	return err
}

func MatchDoctorTypeToUid(c *gin.Context, DoctorId string) (err error) {
	DoctorType := c.GetString("doctor_type")
	uid := c.GetString("uid")
	err = nil

	if DoctorType == "Doctor" && uid != DoctorId {
		err = errors.New("Unauthorized to access this resource")
		return err
	}
	err = CheckDoctorType(c, DoctorType)
	return err
}
