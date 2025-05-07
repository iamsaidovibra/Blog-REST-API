package routes

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/iamsaidovibra/blog-rest-api/database"
	"github.com/iamsaidovibra/blog-rest-api/models"
	"github.com/iamsaidovibra/blog-rest-api/utils"
	"golang.org/x/crypto/bcrypt"
)

type UserSerializer struct {
	ID        uint   `json:"id"`
	FirstName string `json:"first_name" gorm:"not null"`
	LastName  string `json:"last_name" gorm:"not null"`
	Username  string `json:"username" gorm:"uniqueIndex;not null"`
	Email     string `json:"email" gorm:"uniqueIndex;not null"`
	Password  string `json:"-" gorm:"not null"`
	//commmented password out for now
}

func CreateResponseUser(userModel models.User) UserSerializer {
	return UserSerializer{
		ID:        userModel.ID,
		FirstName: userModel.FirstName,
		LastName:  userModel.LastName,
		Username:  userModel.Username,
		Email:     userModel.Email,
		// Password:  userModel.Password,
	}
}

func LoginUser(c *fiber.Ctx) error {
	type LoginInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Find user by email
	var user models.User
	if err := database.Database.Db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	fmt.Println(user.Password)
	fmt.Println(input.Password)
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(input.Password),
	); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Generate token
	token, err := utils.GenerateToken(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not generate token"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user":  CreateResponseUser(user),
	})
}

func CreateUser(c *fiber.Ctx) error {
	var users []models.User
	if err := c.BodyParser(&users); err == nil {
		// Handle bulk creation
		if len(users) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Empty list of users"})
		}

		for i := range users {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(users[i].Password), bcrypt.DefaultCost)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error":   "Could not hash password",
					"details": err.Error(),
				})
			}
			users[i].Password = string(hashedPassword)
		}

		if err := database.Database.Db.Create(&users).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Failed to create users",
				"details": err.Error(),
			})
		}

		responseUsers := make([]UserSerializer, len(users))
		for i, user := range users {
			responseUser := CreateResponseUser(user)
			responseUser.Password = ""
			responseUsers[i] = responseUser
		}
		return c.Status(201).JSON(responseUsers)
	}

	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request format",
			"details": "Expected either:\n" +
				"- Single user object {\"first_name\":...,}\n" +
				"- Array of users [{\"first_name\":...,}, ...]",
		})
	}

	// Hash password for single user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Could not hash password",
			"details": err.Error(),
		})
	}
	user.Password = string(hashedPassword)

	// Handle single user creation
	if err := database.Database.Db.Create(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create user",
			"details": err.Error(),
		})
	}

	// Clear hashed password from response
	user.Password = ""
	return c.Status(201).JSON(CreateResponseUser(user))
}

func GetUsers(c *fiber.Ctx) error {
	// 1) pull the userID out of context
	userID := utils.GetUserID(c)
	// limit, offset := utils.Paginate(c)

	// 2) load exactly that user
	var user models.User
	if err := database.Database.Db.First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	// 3) return your serializer
	return c.Status(200).JSON(CreateResponseUser(user))
}


func findUser(id uint, user *models.User) error {
	database.Database.Db.Find(user, "id = ?", id) //removed & for now
	if user.ID == 0 {
		return errors.New("User doesn't exist")
	}
	return nil
}

func GetUserById(c *fiber.Ctx) error {
	var user models.User
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON("Make sure ID is an integer")
	}

	if err := findUser(uint(id), &user); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	responseUser := CreateResponseUser(user)

	return c.Status(200).JSON(responseUser)

}

func UpdateUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	var user models.User

	if err != nil {
		return c.Status(400).JSON("Make sure id is an integer")
	}

	if err := findUser(uint(id), &user); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	type UpdateUser struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		// Password  string `json:"-"`
	}

	var updateData UpdateUser
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(500).JSON(err.Error())
	}

	user.FirstName = updateData.FirstName
	user.LastName = updateData.LastName
	user.Username = updateData.Username
	user.Email = updateData.Email
	// user.Password = updateData.Password

	database.Database.Db.Save(&user)
	responseUser := CreateResponseUser(user)
	return c.Status(200).JSON(responseUser)

}

func DeleteUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	var user models.User
	if err != nil {
		return c.Status(400).JSON("Make sure ID is an integer")
	}

	if err := findUser(uint(id), &user); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	if err := database.Database.Db.Delete(&user).Error; err != nil {
		return c.Status(404).JSON(err.Error())
	}

	return c.Status(200).SendString("User was DELETED successfully")
}
