ifneq (,$(wildcard ./.env))
	include .env
	export
endif

init:
	terraform init \
		-backend-config="resource_group_name=Terraform" \
		-backend-config="storage_account_name=${TERRAFORM_BACKEND_STORAGE_ACCOUNT}" \
		-backend-config="container_name=tfstate" \
		-backend-config="key=timesheets-sync.tfstate"

validate:
	terraform validate

plan: pack
	terraform plan

apply: pack
	terraform apply -auto-approve

terra: pack
	terraform apply

destroy:
	terraform destroy

pack:
	cd ../ && make pack-fns