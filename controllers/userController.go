package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/akhil/golang-jwt-project/database"
	helper "github.com/akhil/golang-jwt-project/helpers"
	"github.com/akhil/golang-jwt-project/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type isLoggedin struct {
	IsLoggedin bool `json:"isLoggedin"`
}

var DoctorCollection *mongo.Collection = database.OpenCollection(database.Client, "Doctor")
var validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(DoctorPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(DoctorPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("email of password is incorrect")
		check = false
	}
	return check, msg
}

func Signup() gin.HandlerFunc {

	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var Doctor models.Doctor

		if err := c.BindJSON(&Doctor); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(Doctor)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := DoctorCollection.CountDocuments(ctx, bson.M{"email": Doctor.Email})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
		}

		password := HashPassword(*Doctor.Password)
		Doctor.Password = &password

		count, err = DoctorCollection.CountDocuments(ctx, bson.M{"phone": Doctor.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the phone number"})
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email or phone number already exists"})
		}

		Doctor.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		Doctor.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		Doctor.ID = primitive.NewObjectID()
		Doctor.Doctor_id = Doctor.ID.Hex()
		token, refreshToken, _ := helper.GenerateAllTokens(*Doctor.Email, *Doctor.First_name, *Doctor.Last_name, *Doctor.Doctor_type, *&Doctor.Doctor_id)
		Doctor.Token = &token
		Doctor.Refresh_token = &refreshToken

		resultInsertionNumber, insertErr := DoctorCollection.InsertOne(ctx, Doctor)
		if insertErr != nil {
			msg := fmt.Sprintf("Doctor item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, resultInsertionNumber)
	}

}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var Doctor models.Doctor
		var foundDoctor models.Doctor

		if err := c.BindJSON(&Doctor); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := DoctorCollection.FindOne(ctx, bson.M{"email": Doctor.Email}).Decode(&foundDoctor)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is incorrect"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*Doctor.Password, *foundDoctor.Password)
		defer cancel()
		if passwordIsValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		if foundDoctor.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Doctor not found"})
			return
		}

		token, refreshToken, _ := helper.GenerateAllTokens(*foundDoctor.Email, *foundDoctor.First_name, *foundDoctor.Last_name, *foundDoctor.Doctor_type, foundDoctor.Doctor_id)
		helper.UpdateAllTokens(token, refreshToken, foundDoctor.Doctor_id)
		err = DoctorCollection.FindOne(ctx, bson.M{"Doctor_id": foundDoctor.Doctor_id}).Decode(&foundDoctor)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Check the Doctor_type to determine which page to render
		if foundDoctor.Doctor_type != nil && *foundDoctor.Doctor_type == "ADMIN" {
			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"Doctor": foundDoctor,
			})
		}
		if foundDoctor.Doctor_type != nil && *foundDoctor.Doctor_type == "NURSE" {
			c.HTML(http.StatusOK, "nursedashboard.html", gin.H{
				"Nurse": foundDoctor,
			})
		} else {
			c.HTML(http.StatusOK, "home.html", gin.H{
				"Doctor": foundDoctor,
			})
		}
	}
}

func GetDoctors() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckDoctorType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{"$match", bson.D{{}}}}
		groupStage := bson.D{{"$group", bson.D{
			{"_id", bson.D{{"_id", "null"}}},
			{"total_count", bson.D{{"$sum", 1}}},
			{"data", bson.D{{"$push", "$$ROOT"}}}}}}
		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"Doctor_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}}}}}
		result, err := DoctorCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing Doctor items"})
		}
		var allDoctors []bson.M
		if err = result.All(ctx, &allDoctors); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allDoctors[0])
	}
}

func GetDoctor() gin.HandlerFunc {
	return func(c *gin.Context) {
		DoctorId := c.Param("Doctor_id")

		if err := helper.MatchDoctorTypeToUid(c, DoctorId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var Doctor models.Doctor
		err := DoctorCollection.FindOne(ctx, bson.M{"Doctor_id": DoctorId}).Decode(&Doctor)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, Doctor)
	}
}
