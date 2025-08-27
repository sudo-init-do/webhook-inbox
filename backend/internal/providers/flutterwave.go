package providers

// Placeholder: Flutterwave often sends "verif-hash" header = SHA256(secret)
// For v1 we’ll accept if header equals secret; refine later.
func VerifyFlutterwaveSignature(secret, header string) bool {
	return header != "" && header == secret
}
