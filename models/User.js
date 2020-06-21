const mongoose = require("mongoose");

const UserSchema = new mongoose.Schema({
    username: {
        type: String,
        required: true,
    },
    password: {
        type: String,
        required: true
    },
    email: {
        type: String,
        required: true,
    },
    date: {
        type: Date,
        default: Date.now
    },
    blacklisted: {
        type: Boolean,
        required: false
    },
    blacklistedReason: {
        type: String,
        required: false
    },
    ip: {
        type: String,
        required: false
    },
    lastChanged: {
        type: Date,
        required: false
    }
})

module.exports = mongoose.model("users", UserSchema);