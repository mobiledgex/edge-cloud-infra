path "*" {
	capabilities = ["read", "list"]
}

path "+/jwtkeys/*" {
	capabilities = ["create", "read", "update", "delete", "list"]
}

path "+/jwtkeys/*" {
	capabilities = ["create", "read", "update", "delete", "list"]
}

path "jwtkeys/*" {
	capabilities = ["create", "read", "update", "delete", "list"]
}

path "secret/*" {
	capabilities = ["create", "read", "update", "delete", "list"]
}

path "certs/*" {
	capabilities = ["read"]
}

path "auth/github/map/users/+" {
	capabilities = ["create", "read", "update", "delete", "list"]
}
