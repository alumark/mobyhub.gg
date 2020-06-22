const express = require("express");
const bodyParser = require("body-parser");
const mongoose = require("mongoose");
const passport = require("passport");
const path = require('path');
const axios = require('axios');

const users = require("./routes/api/users");

const app = express();

// Bodyparser middleware
app.use(
    bodyParser.urlencoded({
        extended: false
    })
);
app.use(bodyParser.json());
app.use(passport.initialize());
// Passport config
require("./config/passport")(passport);
// Routes
app.use("/api/users", users);

const { mongoURI } = require("./config/keys");

const discordLink = "https://discord.gg/pQrJysn";
app.enable("trust proxy");

app.set('view engine', 'ejs');

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
        res.send(response.data);
    }).catch((error) => {
        console.log(error)
        res.send("Internal server error");
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

const PORT = process.env.PORT || 8080;

if (process.env.NODE_ENV == 'production') {
    app.use(express.static(path.join(__dirname, "client", "build")))
}

app.listen(PORT, () => console.log(`Server up and running on port ${PORT}!`));