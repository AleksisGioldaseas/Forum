.PHONY: run gitDB db-init db-reset db-clean quickstart cleanstart db-populate
DB_FOLDER=data
DB_FILE=data/forum.db
FORUM_SCHEMA=sql/schema.sql
CERTS_DIR=certs
CERT_FILE=$(CERTS_DIR)/cert.pem
KEY_FILE=$(CERTS_DIR)/key.pem

# define load_env
# 	$(eval include .env)
# 	$(eval export $(shell sed 's/=.*//' .env))
# endef

quickstart:
	@$(MAKE) db-init && go run app/main.go

run:
	go run app/main.go

gen-certs:
	@if [ ! -f $(CERT_FILE) ] || [ ! -f $(KEY_FILE) ]; then \
		echo "Generating SSL certificates..."; \
		mkdir -p $(CERTS_DIR); \
		openssl req -x509 -newkey rsa:4096 -keyout $(KEY_FILE) -out $(CERT_FILE) -days 365 -nodes -subj "/CN=localhost"; \
		echo "Certificates generated in $(CERTS_DIR)/"; \
	else \
		echo "SSL Certificates already exist"; \
	fi

clean-certs:
	@if [ -f $(CERT_FILE) ] || [ -f $(KEY_FILE) ]; then \
		echo "Deleting SSL certificates..."; \
		rm -f $(CERT_FILE) $(KEY_FILE); \
		echo "Certificates deleted"; \
	else \
		echo "No certificates found"; \
	fi

gitDB:
	persistence/*
	git commit -m "$(m)"
	git push

db-init:
	@if [ ! -f $(DB_FILE) ]; then \
		echo "Creating database..."; \
		mkdir -p $(DB_FOLDER); \
		mkdir -p $(DB_FOLDER)/images; \
		sqlite3 $(DB_FILE) < $(FORUM_SCHEMA); \
		echo "Database initialized."; \
	else \
		echo "Database already exists."; \
	fi

# WARNING: Deletes all data!
cleanstart:
	@$(MAKE) keystart

# WARNING: Deletes all data!
realcleanstart:
	@$(MAKE) db-reset
	@$(MAKE) db-populate
	# @$(call load_env)
	go run app/main.go

# WARNING: Deletes all data!
keystart:
	@$(MAKE) gen-certs
	@$(MAKE) realcleanstart

db-reset:
	@$(MAKE) db-clean
	@$(MAKE) db-init
	echo "Database reset."

db-clean:
	rm -rf $(DB_FOLDER)
	echo "Database deleted."

db-populate:
	go run app/main.go "populate"

db-test:
	@$(MAKE) db-clean
	@if [ ! -f data/test_forum.db ]; then \
		echo "Creating database..."; \
		mkdir -p data; \
		mkdir -p $(DB_FOLDER)/images; \
		sqlite3 data/test_forum.db < $(FORUM_SCHEMA); \
		echo "Database initialized."; \
	else \
		echo "Database already exists."; \
	fi
	# go test ./persistence/database

