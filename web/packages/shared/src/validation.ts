// Shared input validation helpers used by client and admin frontends.

const USERNAME_PATTERN = /^[a-zA-Z0-9]{2,15}$/

// Characters allowed in the client search bar (Chinese/English, numbers, spaces, common punctuation).
const SEARCH_ALLOWED_CHARS =
  '\\p{Script=Han}\\p{Script=Latin}\\p{Number}\\s.,!?;:\'"()[\\]{}@#&*_\\-~。，、；：？！…—·（）【】《》“”‘’'

/** Sanitizes a live search input by removing invalid characters and capping to 10 chars. */
export function sanitizeSearchInput(value: string): string {
  return value
    .replace(new RegExp(`[^${SEARCH_ALLOWED_CHARS}]`, 'gu'), '')
    .slice(0, 10)
}

/** Returns true if the username (after trim) is exactly 2-15 letters/digits. */
export function isValidUsername(username: string): boolean {
  return USERNAME_PATTERN.test(username.trim())
}

/** Sanitizes a live username input by keeping only letters/digits and capping to 15 chars. */
export function sanitizeUsernameInput(value: string): string {
  return value.replace(/[^a-zA-Z0-9]/g, '').slice(0, 15)
}
