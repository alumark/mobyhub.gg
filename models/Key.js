const mongoose = require("mongoose");

const UserSchema = new mongoose.Schema({
    key: {
        type: String,
        required: true,
    },
    used: {
        type: Boolean,
        required: false
    }
})

module.exports = mongoose.model("keys", UserSchema);