import { writable } from 'svelte/store'
import jwt_decode from "jwt-decode";

// Get the value out of storage on load.
const stored = localStorage.getItem('jwtToken');
// or localStorage.getItem('content')

// Set the stored value or a sane default.
export const jwtToken = writable(stored || null);
export const user = writable(null);

const setUsername = value => {
    try {
        let decoded = jwt_decode(value)
        user.set(decoded.username)
    } catch (e) {
        console.log(`error: ${e}`)
    }
}

// Anytime the store changes, update the local storage value.
jwtToken.subscribe((value) => {
    localStorage.setItem('jwtToken', value)
    setUsername(value);
});