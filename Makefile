.PHONY: run-put, run-delete

generate-certificates:
	openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -sha256 -days 365 --nodes


run-put:
	curl x PUT -d 'Hello, key-values store!' --insecure -v https://localhost:8080/v1/key-a

