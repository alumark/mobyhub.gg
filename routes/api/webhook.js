const express = require("express");
const router = express.Router();

const nodemailer = require("nodemailer");
const crypto = require("crypto");
const nodemailerExpressHandlebars = require("nodemailer-express-handlebars");

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
  return new Promise(async (resolve) => {
    resolve(getRandomString(KEY_LENGTH));
  });
};

const purchased = async order => {
  return new Promise(async (resolve, reject) => {
    if (order.status === "COMPLETED" && order.customer_email) {
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
    } else {
      reject('small issue');
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

function validateShopifySignature() {
    return async (req, res, next) => {
        try {
          const rawBody = req.rawBody
          if (typeof rawBody == 'undefined') {
              throw new Error(
                  'validateShopifySignature: req.rawBody is undefined. Please make sure the raw request body is available as req.rawBody.'
              )
          }

          const hmac = req.headers['x-sellix-signature']
          const hash = crypto
              .createHmac('sha512', process.env.WEBHOOK_SECRET)
              .update(rawBody)
              .digest('hex')

          const signatureOk = crypto.timingSafeEqual(
              Buffer.from(hash),
              Buffer.from(hmac)
          )
          if (!signatureOk) {
              res.status(403)
              res.send('Unauthorized')
              return
          }
            
          next()
      } catch (err) {
          next(err)
      }
    }
}

router.post(
    '/purchased',
    validateShopifySignature(),
    async (req, res) => {
        const key = await purchased(req.body.data).catch(() => {
          res.status(400).json({
            message: "authentication failed"
          })
        });

        console.log(`new key ${key}`);
        res.status(200).json({
          message: "ok"
        })
    }
)

module.exports = router;
