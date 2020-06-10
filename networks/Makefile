########################################
### These targets were broken out of the main Makefile to enable easy setup of testnets.
### They use a form of terraform + ansible to build full nodes in AWS.
### The shell scripts in this folder are example uses of the targets.

# Name of the testnet. Used in chain-id.
TESTNET_NAME?=remotenet

# Name of the servers grouped together for management purposes. Used in tagging the servers in the cloud.
CLUSTER_NAME?=$(TESTNET_NAME)

# Number of servers to put in one availability zone in AWS.
SERVERS?=1

# Number of regions to use in AWS. One region usually contains 2-3 availability zones.
REGION_LIMIT?=1

# Path to nodef for deployment. Must be a Linux binary.
BINARY?=$(CURDIR)/../build/nodef
CLIF_BINARY?=$(CURDIR)/../build/clif
EE_BINARIY?=$(CURDIR)/../build/CasperLabs/

# Path to the genesis.json and config.toml files to deploy on full nodes.
GENESISFILE?=$(CURDIR)/../build/genesis.json
CONFIGFILE?=$(CURDIR)/../build/config.toml

MANIFEST?=$(CURDIR)/../x/executionlayer/resources/manifest.toml
CONSENSUS_MODULE ?= friday

# Name of application for app deployments
APP_NAME ?= faucettestnet1
# Region to deploy VPC and application in AWS
REGION ?= us-east-2

SSH_PRIVATE_FILE?=$(HOME)/workspaces/AWS/friday_consensus.pem
SSH_PUBLIC_FILE?=$(HOME)/workspaces/AWS/friday_consensus.pub

all:
	@echo "There is no all. Only sum of the ones."

disclaimer:
	@echo "WARNING: These are example network configuration scripts only and have not undergone security review. They should not be used for production deployments."

########################################
### Extract genesis.json and config.toml from a node in a cluster

extract-config: disclaimer
	#Make sure you have AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY or your IAM roles set for AWS API access.
	@if ! [ -f $(SSH_PUBLIC_FILE) ]; then ssh-keygen ; fi
	cd remote/ansible && \
	ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook \
		-i inventory/ec2.py \
		-l "tag_Environment_$(CLUSTER_NAME)" \
		-b -u ec2-user \
		-e TESTNET_NAME="$(TESTNET_NAME)" \
		-e GENESISFILE="$(GENESISFILE)" \
		-e CONFIGFILE="$(CONFIGFILE)" \
		extract-config.yml


########################################
### Remote validator nodes using terraform and ansible in AWS

validators-start: disclaimer
	#Make sure you have AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY or your IAM roles set for AWS API access.
	@if ! [ -f $(SSH_PUBLIC_FILE) ]; then ssh-keygen ; fi
	@if [ -z "`file $(BINARY) | grep 'ELF 64-bit'`" ]; then echo "Please build a linux binary using 'make build-linux'." ; false ; fi
	cd remote/terraform-aws && terraform init && (terraform workspace new "$(CLUSTER_NAME)" || terraform workspace select "$(CLUSTER_NAME)") && terraform apply -auto-approve -var SSH_PUBLIC_FILE="$(SSH_PUBLIC_FILE)" -var SSH_PRIVATE_FILE="$(SSH_PRIVATE_FILE)" -var TESTNET_NAME="$(CLUSTER_NAME)" -var SERVERS="$(SERVERS)" -var REGION_LIMIT="$(REGION_LIMIT)"
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/ec2.py -l "tag_Environment_$(CLUSTER_NAME)" -u ec2-user -b -e BINARY=$(BINARY) -e CLIF_BINARY=$(CLIF_BINARY) -e TESTNET_NAME="$(TESTNET_NAME)" add-lcd.yml
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/ec2.py -v -l "tag_Environment_$(CLUSTER_NAME)" -u ec2-user -b -e BINARY=$(BINARY) -e CLIF_BINARY=$(CLIF_BINARY) -e EE_BINARIY=$(EE_BINARIY) -e MANIFEST=$(MANIFEST) -e CONSENSUS_MODULE=$(CONSENSUS_MODULE) -e TESTNET_NAME="$(TESTNET_NAME)" setup-validators.yml
	cd remote/ansible && ansible-playbook -i inventory/ec2.py -v -l "tag_Environment_$(CLUSTER_NAME)" -u ec2-user -b start.yml

validators-stop: disclaimer
	cd remote/terraform-aws && terraform workspace select "$(CLUSTER_NAME)" && terraform destroy -force -var SSH_PUBLIC_FILE="$(SSH_PUBLIC_FILE)" -var SSH_PRIVATE_FILE="$(SSH_PRIVATE_FILE)" && terraform workspace select default && terraform workspace delete "$(CLUSTER_NAME)"
	rm -rf remote/ansible/keys/ remote/ansible/files/

validators-status: disclaimer
	cd remote/ansible && ansible-playbook -i inventory/ec2.py -l "tag_Environment_$(CLUSTER_NAME)" status.yml

#validators-clear:
#	cd remote/ansible && ansible-playbook -i inventory/ec2.py -l "tag_Environment_$(CLUSTER_NAME)" -u centos -b clear-config.yml


# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: all extract-config validators-start validators-stop validators-status
