const express = require("express");
const bodyParser = require("body-parser");
const mongoose = require("mongoose");
const passport = require("passport");
<<<<<<< HEAD
const ejs = require("ejs");
=======
>>>>>>> e1e8af15a2767b0371a5ecc92fefb3eee159ac6c

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
app.use(express.static('client/build'))

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

const PORT = process.env.PORT | 5000;

app.listen(PORT, () => console.log(`Server up and running on port ${PORT}!`));