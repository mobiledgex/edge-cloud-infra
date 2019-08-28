path "secret/data/ansible/*" {
	capabilities = ["read", "list"]
}

path "secret/metadata/ansible/*" {
	capabilities = ["list"]
}

path "secret/data/ansible/main/*" {
	capabilities = []
}

path "secret/metadata/ansible/main/*" {
	capabilities = []
}

path "secret/data/ansible/prod/*" {
	capabilities = []
}

path "secret/metadata/ansible/prod/*" {
	capabilities = []
}

path "secret/data/ansible/internal/*" {
	capabilities = []
}

path "secret/metadata/ansible/internal/*" {
	capabilities = []
}
