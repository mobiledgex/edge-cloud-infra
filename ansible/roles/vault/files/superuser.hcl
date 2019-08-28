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

# Override limitations set in the github-dev policy
path "secret/data/ansible/*" {
	capabilities = ["create", "read", "update", "delete", "list"]
}

path "secret/metadata/ansible/*" {
	capabilities = ["create", "read", "update", "delete", "list"]
}

path "secret/data/ansible/main/*" {
	capabilities = ["create", "read", "update", "delete", "list"]
}

path "secret/metadata/ansible/main/*" {
	capabilities = ["create", "read", "update", "delete", "list"]
}

path "secret/data/ansible/prod/*" {
	capabilities = ["create", "read", "update", "delete", "list"]
}

path "secret/metadata/ansible/prod/*" {
	capabilities = ["create", "read", "update", "delete", "list"]
}

path "secret/data/ansible/internal/*" {
	capabilities = ["create", "read", "update", "delete", "list"]
}

path "secret/metadata/ansible/internal/*" {
	capabilities = ["create", "read", "update", "delete", "list"]
}
