import { writable } from 'svelte/store'
import jwt_decode from "jwt-decode";

// Get the value out of storage on load.
const stored = localStorage.getItem('jwtToken');
// or localStorage.getItem('content')

console.log(stored)

// Set the stored value or a sane default.
export const jwtToken = writable(stored || null);
export const user = writable(null);

// Anytime the store changes, update the local storage value.
jwtToken.subscribe((value) => {
    localStorage.setItem('jwtToken', value)
    let decoded = jwt_decode(value)
    user.set(decoded.username)
});