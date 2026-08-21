// Shared input validation helpers used by client and admin frontends.

// Characters allowed in the client search bar (Chinese/English, numbers, spaces, common punctuation).
const SEARCH_ALLOWED_CHARS =
  '\\p{Script=Han}\\p{Script=Latin}\\p{Number}\\s.,!?;:\'"()[\\]{}@#&*_\\-~。，、；：？！…—·（）【】《》“”‘’'

/** Sanitizes a live search input by removing invalid characters and capping to 10 chars. */
export function sanitizeSearchInput(value: string): string {
  return value.replace(new RegExp(`[^${SEARCH_ALLOWED_CHARS}]`, 'gu'), '').slice(0, 10)
}

// Simple RFC 5322 compatible email regex; sufficient for client-side sanity check.
// Backend enforces strict format via gin validator `email` tag.
const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

/** Returns true if the email (after trim) matches a basic email shape. */
export function isValidEmail(email: string): boolean {
  return EMAIL_PATTERN.test(email.trim())
}

/** Normalizes email input: trim + lowercase + cap to 128 chars (matches backend max). */
export function sanitizeEmailInput(value: string): string {
  return value.trim().toLowerCase().slice(0, 128)
}
