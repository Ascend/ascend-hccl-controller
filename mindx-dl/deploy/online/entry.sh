#!/bin/bash
set -e
ansible-playbook docker_install.yaml
ansible-playbook install.yaml
ansible-playbook deploy.yaml