const express = require("express");
const router = express.Router();
const bcrypt = require("bcryptjs");
const jwt = require("jsonwebtoken");
const keys = require("../../config/keys");
const axios = require("axios");
const nodemailer = require("nodemailer");
const crypto = require("crypto");
const nodemailerExpressHandlebars = require("nodemailer-express-handlebars");
// Load input validation
const validateRegisterInput = require("../../validation/register");
const validateLoginInput = require("../../validation/login");
// Load User model
const User = require("../../models/User");
const Key = require("../../models/Key")
const { emailUsername, emailPassword, token } = require("../../config/keys");

function getRandomString(length) {
  var randomChars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz01234567891234567890!?-.';
  var result = '';
  for ( var i = 0; i < length; i++ ) {
      result += randomChars.charAt(Math.floor(Math.random() * randomChars.length));
  }
  return result;
}

const createKey = async () => {
  return new Promise(async (resolve, reject) => {
    resolve(getRandomString(KEY_LENGTH));
  });
};

module.exports.purchased = async order => {
  return new Promise(async (resolve, reject) => {
  
    if (order.status === 1 && order.customer_email) {
      let transporter = nodemailer.createTransport({
        service: 'gmail',
        auth: {
          user: process.env.EMAIL_ADDRESS,
          pass: process.env.EMAIL_PASSWORD, // naturally, replace both with your real credentials or an application-specific password
        },
        tls: {
          rejectUnauthorized: false
        }
      });

      const key = await createKey();

      if (key) {
        let info = await transporter.sendMail({
          from: `mobyhub <${process.env.EMAIL_ADDRESS}>`,
          to: order.customer_email,
          subject: "About",
          text: `Your key: ${key}, enter it at https://mobyhub-pipeline.glitch.me/signup/`,
          html: `<b>Your key: ${key}, enter it </b><a href="https://mobyhub-pipeline.glitch.me/signup/">here</a>`
        });

        let newKey = new Key({ key: key });
        await newKey.save((err) => {
          if (err) console.log(err);
          console.log(`saved key! ${key}`);
        });

        resolve(key);
      } else {
        reject("no key")
      }
    }
  });
};

let transporter = nodemailer.createTransport({
    service: "gmail",
    auth: {
      user: emailUsername,
      pass: emailPassword // naturally, replace both with your real credentials or an application-specific password
    },
    tls: {
      rejectUnauthorized: false
    }
  });

  transporter.use(
    "compile",
    nodemailerExpressHandlebars({
      viewEngine: {
        extName: ".hbs",
        partialsDir: __dirname + "/views",
        layoutsDir: __dirname + "/views",
        defaultLayout: "email.hbs"
      },
      viewPath: __dirname + "/views",
      extName: ".hbs"
    })
  );

// @route POST api/users/register
// @desc Register user
// @access Public
router.post("/register", (req, res) => {
    // Form validation
    const { errors, isValid } = validateRegisterInput(req.body);
    if (!isValid) {
        return res.status(400).json(errors);
    };

    const { username, password, key, email } = req.body;
    Key.findOne({ key }, (err, newKey) => {
        if (newKey && !newKey.used) {
            Key.updateOne({ key }, { used: true }, err => {
                if (err) {
                    console.log(err);
                    return res.status(500).json({
                        message: "Internal server error"
                    });
                }
                User.findOne({ username: username }, (err, user) => {
                    if (err) {
                        console.log(err);
                        return res.status(500).json({
                            message: "Internal server error"
                        });
                    };

                    if (!user) {
                        let user = new User({
                            username: username,
                            password: password,
                            email: email
                        });

                        bcrypt.genSalt(10, (err, salt) => {
                            if (err) {
                                console.log(err);
                                return res.status(500).json({
                                    message: "Internal server error"
                                });
                            };

                            bcrypt.hash(user.password, salt, (err, hash) => {
                                if (err) {
                                    console.log(err);
                                    return res.status(500).json({
                                        message: "Internal server error"
                                    });
                                };

                                user.password = hash;
                                user
                                    .save()
                                    .then(() => {
                                        res.json(user)
                                    })
                                    .catch(err => {
                                        return res.status(500).json({
                                            message: "Internal server error"
                                        });
                                    });
                            });
                        });
                    } else {
                        return res.status(400).json({
                            username: "Username already exists!"
                        });
                    };
                });
            });
        } else if (newKey && newKey.used) {
            return res.status(400).json({
                key: "Key has already been used!"
            });
        } else {
            return res.status(400).json({
                key: "Invalid key!"
            });
        };
    });
  });

  // @route POST api/users/login
// @desc Login user and return JWT token
// @access Public
router.post("/login", (req, res) => {
    // Form validation
    const { errors, isValid } = validateLoginInput(req.body);
  // Check validation
    if (!isValid) {
        return res.status(400).json(errors);
    }
    const username = req.body.username;
    const password = req.body.password;
        // Find user by email
    User.findOne({ username }).then(user => {
        // Check if user exists
        if (!user) {
            return res.status(404).json({ username: "Username not found" });
        }
        if (user.blacklisted) {
          return res.status(403).json({ username: `You have been blacklisted: "${user.blacklistedReason}"` })
        }
        bcrypt.compare(password, user.password).then(isMatch => {
            if (isMatch) {
                // User matched
                // Create JWT Payload
                const payload = {
                    id: user.id,
                    username: user.username
                };
                jwt.sign(
                    payload,
                    keys.secretOrKey,
                    {
                        expiresIn: 31556926 // 1 year in seconds
                    },
                    (err, token) => {
                        res.json({
                        success: true,
                        token: "Bearer " + token
                        });
                    }
                );
            } else {
                return res
                    .status(400)
                    .json({ passwordincorrect: "Password incorrect" });
            }
        });
    });
});

function checkAuthentication(username, password, ip) {
    return new Promise((resolve, reject) => {
      User.findOne({ username: username }, (err, newHash) => {
        if (err) {
          console.log(err);
          reject("Internal server error");
          return;
        }
  
        if (newHash) {
          if (newHash.blacklisted) {
            reject(`You are blacklisted for: \n${newHash.blacklistedReason}.`);
            return;
          }
  
          bcrypt.compare(password, newHash.password, async function(err, result) {
            console.log(result);
            if (err) {
              console.log(err);
              reject("Internal server error");
              return;
            }
            if (result) {
              bcrypt.compare(ip, newHash.ip, async function(err, success) {
                if (success) {
                  resolve();
                } else if (
                  (!success && !newHash.lastChanged) ||
                  new Date().getTime() - newHash.lastChanged.getTime() >= TIME
                ) {
                  User.updateOne(
                    { username: username },
                    { ip: await generateHash(ip), lastChanged: new Date() },
                    function(err, res) {
                      if (err) {
                        console.log(err);
                        reject("Internal server error");
                        return;
                      }
                      resolve();
                    }
                  );
                } else if (
                  !success &&
                  (newHash.lastChanged &&
                    new Date().getTime() - newHash.lastChanged.getTime() <= TIME)
                ) {
                  reject(
                    `Your IP was changed too recently.  Please wait at least ${moment
                      .duration(
                        new Date().getTime() - newHash.lastChanged.getTime()
                      )
                      .format("mm minutes, ss seconds")}.`
                  );
                }
              });
            } else {
              reject("Invalid password");
            }
          });
        } else {
          reject("Invalid username");
        }
      });
    });
  };
  
router.post("/script", (req, res) => {
    let { username, password } = req.body;
    let ip = req.ip;

    username = username.replace(/\s+/g, "");

    checkAuthentication(username, password, ip)
      .then(() => {
          axios({
              method: "get",
              url: "https://raw.githubusercontent.com/alumark/mobyhub/master/init.lua",
              headers: {
                  Authorization: "token " + token
              }
          }).then(response => {
              return res.send(response.data);
          }).catch(() => {
              res.status(500);
          });
      }).catch(errorMessage => {
          return res.status(400).json({
              status: "error",
              message: errorMessage,
          })
      });
    })

router.post("/purchase", async (req, res) => {
  const signature = crypto
    .createHmac("SHA256", process.env.WEBHOOK_SECRET)
    .digest("hex");

  /*if (signature === req.headers["x-sellix-signature"]) {*/
    let key = await purchased(req.body.data);
    res.send({
      success: "ok"
    });
  /*} else {
    res.send({
      success: "error",
      message: "authentication failed"
    });
  }*/
});

  module.exports = router;