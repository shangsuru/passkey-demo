export function isValidEmail(email: string): boolean {
  const emailRegex = /^.+@.+\..+$/;
  return email !== "" && emailRegex.test(email);
}
