package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	realip "github.com/krecu/fasthttp-realip"
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
	IP          string             `bson:"ip,omitempty"`
	LastChanged primitive.DateTime `bson:"lastChanged,omitempty"`
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

type JSONData struct {
	Email string `json:"customer_email"`
}

var rxEmail = regexp.MustCompile(".+@.+\\..+")

var (
	github       string
	mongo_client *mongo.Client
	users        *mongo.Collection
	keys         *mongo.Collection
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz01234567891234567890!?-."

func generate_key(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}

func send(email string, body string, uniqid string) {
	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	to := email

	msg := "from: mobyhub <" + from + ">\n" +
		"to: " + to + "\n" +
		"Subject: mobyhub order - " + uniqid + "\n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, password, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

	log.Printf("sent")
}

func fmtDuration(time time.Duration) string {
	seconds := math.Round(time.Seconds())
	minutes := int(math.Floor(seconds / 60))
	remainingSeconds := int(seconds) % 60
	return fmt.Sprintf("%d minutes %d seconds", minutes, remainingSeconds)
}

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Printf("Error loading .env file")
	}

	mongo_client, err = LoadMongoDriver()
	users = mongo_client.Database("Cluster0").Collection("users")
	keys = mongo_client.Database("Cluster0").Collection("keys")

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
			fmt.Println(err.Error())
			return c.Status(400).JSON(&fiber.Map{
				"password": err.Error(),
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
		http_client := http.Client{}
		req, err := http.NewRequest("GET", "https://raw.githubusercontent.com/alumark/mobyhub/master/login.lua", nil)

		if err != nil {
			return err
		}

		req.Header = http.Header{
			"Authorization": {fmt.Sprintf("token %s", github)},
		}

		res, err := http_client.Do(req)
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

		ip := realip.FromRequest(c.Context()) // dw this is gonna be encrypted
		hashedIP, err := bcrypt.GenerateFromPassword([]byte(ip), bcrypt.DefaultCost)

		if err != nil {
			log.Println(err.Error())
			return err
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.IP), []byte(ip)); err != nil {
			log.Println(err.Error())
			if time.Now().Sub(user.LastChanged.Time()) <= time.Hour/4 {
				comment := user.LastChanged.Time().Add(time.Hour / 4).Sub(time.Now())
				return c.Status(403).JSON(&fiber.Map{
					"password": fmt.Sprintf("IP Changed too recently, please wait: %s", fmtDuration(comment)),
				})
			} else {
				result := users.FindOneAndUpdate(context.TODO(), bson.M{"username": username}, bson.M{"$set": bson.M{"lastChanged": time.Now(), "ip": string(hashedIP)}})
				if result.Err() != nil {
					log.Panic(result.Err().Error())
					return result.Err()
				}
			}
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

		final := obfuscate(string(body))

		return c.SendString(string(final))
	})

	app.Post("/api/webhook/purchased", func(c *fiber.Ctx) error {
		payload := struct {
			Event string `json:"event"`
			Data  struct {
				Email  string `json:"customer_email"`
				Uniqid string `json:"uniqid"`
			} `json:"data"`
		}{}

		if err := c.BodyParser(&payload); err != nil {
			return err
		}

		fmt.Printf("%+v\n", payload)

		hash := hmac.New(sha512.New, []byte(os.Getenv("WEBHOOK_SECRET")))
		hash.Write(c.Body())
		final_hash := hex.EncodeToString(hash.Sum(nil))

		fmt.Printf("%s %s", final_hash, c.Get("X-Sellix-Signature"))

		if final_hash != c.Get("X-Sellix-Signature") {
			return c.Status(403).JSON(&fiber.Map{
				"status": "unauthorized",
			})
		}

		key := generate_key(24, charset)
		send(payload.Data.Email, "Your key: "+key+" enter it at https://mobyhub.gg", payload.Data.Uniqid)

		_, err := keys.InsertOne(context.TODO(), bson.M{"key": key})
		if err != nil {
			return err
		}

		return c.Status(200).JSON(&fiber.Map{
			"status": "ok",
		})

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

func obfuscate(script string) string {
	id := generate_key(12, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	newpath := filepath.Join(".", id)
	err := os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		log.Print(err.Error())
	}
	f, err := os.Create(filepath.Join(".", id, "in.lua"))
	if err != nil {
		log.Print(err.Error())
	}

	_, err2 := f.WriteString(script)
	if err2 != nil {
		log.Print(err.Error())
	}

	f.Close()

	ctx := context.Background()
	cli, err := client.NewEnvClient(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Print(err.Error())
	}

	out, err := cli.ImagePull(ctx, "docker.io/library/moaufmklo", types.ImagePullOptions{})
	if err != nil {
		//log.Panic(err)
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "docker-ironbrew",
	}, &container.HostConfig{
		AutoRemove: true,
		Binds: []string{
			fmt.Sprintf("%s/:/data", newpath),
		},
	}, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	content, err := os.ReadFile(filepath.Join(".", id, "out.lua"))

	if err != nil {
		log.Print(err.Error())
	}

	if content == nil {
		return string("bruh")
	}

	return string(content)
}

func LoadMongoDriver() (*mongo.Client, error) {
	mongo_client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGO_URI")))

	if err != nil {
		return mongo_client, err
	}

	if err := mongo_client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return mongo_client, err
	}

	return mongo_client, nil
}
