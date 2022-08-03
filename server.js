const express = require("express");
const bodyParser = require("body-parser");
const mongoose = require("mongoose");
const path = require("path")
const axios = require("axios")
const passport = require("passport");

const users = require("./routes/api/users");
const webhook = require("./routes/api/webhook");

const app = express();

// Bodyparser middleware
app.use(
    bodyParser.urlencoded({
        extended: false
    })
);
app.use(
    express.json({
        limit: '50mb',
        verify: (req, res, buf) => {
            req.rawBody = buf
        }
    })
)
app.use(passport.initialize());
// Passport config
require("./config/passport")(passport);
// Routes
app.use("/api/users", users);
app.use("/api/webhook", webhook);

const { mongoURI } = require("./config/keys");

const discordLink = "https://discord.gg/pQrJysn";
app.enable("trust proxy");

app.get("/discord", (req, res) => {
    res.send(discordLink);
});

app.get("/loadstring", (req, res) => {
    axios({
        method: "get",
        url: "https://raw.githubusercontent.com/alumark/mobyhub/master/login.lua",
        headers: {
            Authorization: "token " + process.env.TOKEN
        }
    }).then(response => {
        console.log("succesfully got script")
        return res.send(response.data);
    }).catch(() => {
        res.status(500);
    });
});

mongoose
    .connect(mongoURI,
            {
                useNewUrlParser: true,
                useUnifiedTopology: true
            }
        )
        .then(() => console.log("Successfully connected to MongoDB database!"))
        .catch((error) => console.log(error));

const PORT = process.env.PORT || 80;

if (process.env.NODE_ENV == 'production') {
    app.use(express.static(path.join(__dirname, "client", "dist")))
} else {
    require('dotenv').config()
    app.use(function(req, res, next) {
        res.header("Access-Control-Allow-Origin", "*"); // update to match the domain you will make the request from
        next();
    });
}

app.listen(PORT, () => console.log(`Server up and running on port ${PORT}!`));
