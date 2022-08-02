const Validator = require("validator");
const isEmpty = require("is-empty");

module.exports = function validateRegisterInput(data) {
    let errors = {}

    data.username = !isEmpty(data.username) ? data.username : "";
    data.email = !isEmpty(data.email) ? data.email : "";
    data.password = !isEmpty(data.password) ? data.password : "";
    data.password2 = !isEmpty(data.password2) ? data.password2 : "";
    data.key = !isEmpty(data.key) ? data.key : "";

    if (Validator.isEmpty(data.username)) {
        errors.username = "Name field is required!";
    } else if (!Validator.isEmail(data.email)) {
        errors.email = "Email is not valid!"
    }

    if (Validator.isEmpty(data.password)) {
        errors.password = "Password field is required!"
    }

    if (Validator.isEmpty(data.password2)) {
        errors.password = "Please confirm your password."
    }

    if (!Validator.isLength(data.username, { min: 3, max: 32 })) {
        errors.username = "Username must be at least 3 characters.";
    }

    if (!Validator.isLength(data.password, { min: 6, max: 84 })) {
        errors.password = "Password must be at least 6 characters.";
    }

    if (!Validator.equals(data.password, data.password2)) {
        errors.password2 = "Password does not match."
    }

    if (!Validator.isAlphanumeric(data.username, 'en-US')) {
        errors.username = "Username must contain latin letters and numbers."
    }

    return {
        errors,
        isValid: isEmpty(errors)
    }
}