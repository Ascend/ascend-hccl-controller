#!/bin/bash
set -e
ansible-playbook -vv docker_install.yaml
ansible-playbook -vv install.yaml
ansible-playbook -vv deploy.yaml