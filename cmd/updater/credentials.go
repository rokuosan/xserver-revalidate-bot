package main

func maskCredential(credential string) string {
	if len(credential) <= 4 {
		return "****"
	}
	return credential[:2] + "****" + credential[len(credential)-2:]
}
