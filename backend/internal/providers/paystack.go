package providers

// Placeholder: Paystack often sends "x-paystack-signature" = HMAC SHA512(reqBody, secret)
// Implement later; accept non-empty for v1.
func VerifyPaystackSignature(_ string, header string) bool {
	return header != ""
}
