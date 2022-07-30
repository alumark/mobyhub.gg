const express = require("express");
const bodyParser = require("body-parser");
const mongoose = require("mongoose");
const path = require("path")
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
app.use(bodyParser.json());
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

mongoose
    .connect(mongoURI,
            {
                useNewUrlParser: true,
                useUnifiedTopology: true
            }
        )
        .then(() => console.log("Successfully connected to MongoDB database!"))
        .catch((error) => console.log(error));

const PORT = process.env.PORT || 5000;

if (process.env.NODE_ENV == 'production') {
    app.use(express.static(path.join(__dirname, "client", "build")))
}

app.listen(PORT, () => console.log(`Server up and running on port ${PORT}!`));
