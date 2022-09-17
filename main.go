package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Username    string             `bson:"username,omitempty"`
	Password    string             `bson:"password,omitempty"`
	Blacklisted bool               `bson:"blacklisted,omitempty"`
	Reason      string             `bson:"reason,omitempty"`
}

type Key struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Key  string             `bson:"key,omitempty"`
	Used bool               `bson:"used,omitempty"`
}

type Error struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterError struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Password2 string `json:"password2"`
	Key       string `json:"key"`
	Email     string `json:"email"`
}

var rxEmail = regexp.MustCompile(".+@.+\\..+")

var (
	github string
	client *mongo.Client
	users  *mongo.Collection
	keys   *mongo.Collection
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Printf("Error loading .env file")
	}

	client, err = LoadMongoDriver()
	users = client.Database("Cluster0").Collection("users")
	keys = client.Database("Cluster0").Collection("keys")

	if err != nil {
		panic(err)
	}

	github = os.Getenv("TOKEN")

	app := fiber.New()

	discordLink := "https://discord.gg/pQrJysn"

	app.Get("/discord", func(c *fiber.Ctx) error {
		return c.SendString(discordLink)
	})

	app.Post("/api/users/login", func(c *fiber.Ctx) error {
		payload := struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{}

		if err := c.BodyParser(&payload); err != nil {
			fmt.Println(err)
			return err
		}

		valid := Error{}

		if payload.Username == "" || payload.Password == "" {
			return c.Status(400).JSON(valid)
		}

		var user User
		if err := users.FindOne(context.TODO(), bson.M{"username": payload.Username}).Decode(&user); err != nil {
			valid.Username = "User does not exist"
			return c.Status(400).JSON(valid)
		}

		if err := CheckAuthentication(user, []byte(payload.Password)); err != nil {
			return c.Status(400).JSON(&fiber.Map{
				"password": err,
			})
		}

		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)

		claims["id"] = user.ID
		claims["username"] = user.Username
		claims["exp"] = time.Now().Add(31556926).Unix()

		tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_OR_KEY")))

		if err != nil {
			return err
		}

		return c.Status(200).JSON(&fiber.Map{
			"success": true,
			"token":   fmt.Sprintf("Bearer %s", tokenString),
		})

	})

	app.Get("/loadstring", func(c *fiber.Ctx) error {
		client := http.Client{}
		req, err := http.NewRequest("GET", "https://raw.githubusercontent.com/alumark/mobyhub/master/login.lua", nil)

		if err != nil {
			return err
		}

		req.Header = http.Header{
			"Authorization": {fmt.Sprintf("token %s", github)},
		}

		res, err := client.Do(req)
		if err != nil {
			return err
		}

		body, err := ioutil.ReadAll(res.Body)

		if err != nil {
			return err
		}

		return c.SendString(string(body))
	})

	app.Get("/api/users/script/:username/:password", func(c *fiber.Ctx) error {
		username := c.Params("username")
		password := c.Params("password")

		var user User
		if err := users.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user); err != nil {
			return c.Status(400).JSON(&fiber.Map{
				"username": "User does not exist!",
			})
		}

		if err := CheckAuthentication(user, []byte(password)); err != nil {
			fmt.Printf(err.Error())
			return c.Status(400).JSON(&fiber.Map{
				"password": err.Error(),
			})
		}

		client := http.Client{}
		req, err := http.NewRequest("GET", "https://raw.githubusercontent.com/alumark/mobyhub/master/init.lua", nil)

		if err != nil {
			return err
		}

		req.Header = http.Header{
			"Authorization": {fmt.Sprintf("token %s", github)},
		}

		res, err := client.Do(req)
		if err != nil {
			return err
		}

		body, err := ioutil.ReadAll(res.Body)

		if err != nil {
			return err
		}

		return c.SendString(string(body))
	})

	app.Post("/api/users/register", func(c *fiber.Ctx) error {
		payload := struct {
			Username  string `json:"username"`
			Password  string `json:"password"`
			Password2 string `json:"password2"`
			Key       string `json:"key"`
			Email     string `json:"email"`
		}{}

		if err := c.BodyParser(&payload); err != nil {
			return err
		}

		valid := RegisterError{}

		if payload.Password == "" {
			valid.Password = "Password cannot be empty"
		}

		if payload.Username == "" {
			valid.Username = "Please enter a valid username"
		}

		if len(payload.Username) < 3 {
			valid.Username = "Username must be at least 3 characters long."
		}

		if len(payload.Username) > 32 {
			valid.Username = "Username cannot exceed 32 character.s"
		}

		if is_alphanumeric := regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(valid.Username); is_alphanumeric != true {
			valid.Username = "Username must contain only latin characters and numbers"
		}

		if payload.Password2 == "" {
			valid.Password2 = "Password cannot be empty."
		}

		if payload.Password != payload.Password2 {
			valid.Password2 = "Password does not match."
		}

		var user User
		if err := users.FindOne(context.TODO(), bson.M{"username": payload.Username}).Decode(&user); err == nil {
			valid.Username = "A user already exists with that name!"
		}

		var key Key
		if err := users.FindOne(context.TODO(), bson.M{"key": payload.Key}).Decode(&key); err != nil {
			valid.Key = "Please enter a valid key."
		}

		match := rxEmail.Match([]byte(payload.Email))
		if match == false {
			valid.Email = "Please enter a valid email!"
		}

		if valid.Email != "" || valid.Username != "" || valid.Key != "" || valid.Password != "" || valid.Password2 != "" {
			return c.Status(400).JSON(valid)
		}

		if key.Used == true {
			valid.Key = "Key has already been used!"
			return c.Status(400).JSON(valid)
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)

		if err != nil {
			return err
		}

		result, err := users.InsertOne(context.TODO(), bson.D{{"username", payload.Username}, {"password", string(hashed)}, {"email", payload.Email}})

		if err != nil {
			return err
		}

		fmt.Printf("Success!: %s", result)

		return c.Status(200).JSON(&fiber.Map{
			"success": true,
		})
	})

	app.Static("/", "./client/dist")

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	app.Listen(fmt.Sprintf(":%v", port))
}

func CheckAuthentication(user User, password []byte) error {
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), password); err != nil {
		return fmt.Errorf("Invalid password")
	} else {
		if user.Blacklisted != false {
			reason := user.Reason
			if reason == "" {
				reason = "You have been blacklisted"
			}
			return fmt.Errorf("You are blacklisted: %s", reason)
		}

		return nil
	}
}

func LoadMongoDriver() (*mongo.Client, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGO_URI")))

	if err != nil {
		return client, err
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return client, err
	}

	return client, nil
}
