.PHONY: run-put, run-delete

run-put:
	curl x PUT -d 'Hello, key-values store!' --insecure -v https://localhost:8080/v1/key-a

