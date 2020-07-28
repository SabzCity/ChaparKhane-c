#!/bin/bash
# title:         install.sh
# description:   Clone|build|run apps
# author:        Omid Hekayati
# created:       Oct 22 2016
# updated:       Jun 22 2020
# version:       0.3
# usage:         ./install.sh || bash install.sh
#==============================================================================

# var's
BLUE='\033[0;34m'
GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
YELLOW='\033[0;33m'
NC='\033[0m'
codeUser=sabzcity
repo_Dir="${HOME}/sabz.city"


# Welcoming message
printf "${CYAN}
                                                        *************
                                                        Welcome back!
                                                        *************
       ${NC}\n"

# Install||Update needed apps
printf "\n${GREEN}Install||Update needed apps\n"
apt-get -y install software-properties-common || true
add-apt-repository ppa:longsleep/golang-backports || true
add-apt-repository ppa:certbot/certbot || true
apt-get update
apt-get -y upgrade
apt-get -y install golang-go git certbot

# Get|Update git repo
printf "\n${GREEN}Clone||Pull repo to ${GREEN} $repo_Dir/ ${NC}\n"
git clone https://github.com/SabzCity/sabz.city.git $repo_Dir/ --recursive --shallow-submodules || true
git -C $repo_Dir/ checkout -f || true
git -C $repo_Dir/ pull --recurse-submodules || true

# Clone||Pull modules of sabz.city repo

# Get||Renew certificate
certbot certonly --standalone --preferred-challenges http --domain sabz.city,www.sabz.city \
--csr $repo_Dir/secret/sabz.city.csr --cert-path $repo_Dir/secret/sabz.city.crt --key-path $repo_Dir/secret/sabz.city-private.key \
--fullchain-path $repo_Dir/secret/sabz.city-fullchain.crt --chain-path $repo_Dir/secret/sabz.city-chain.crt

# Get repo dependecies & Build app
printf "\n${GREEN}build sabz.city app ${NC}\n"
cd $repo_Dir
go get
go build

# Install||Update systemd .service files
printf "\n${GREEN}Update sabz.city.service Systemd ${NC}\n"
cp $repo_Dir/sabz.city.service /lib/systemd/system/ || true
systemctl enable sabz.city.service || true

# update systemd daemon
printf "\n${GREEN}Update systemd daemon ${NC}\n"
systemctl daemon-reload

# Restart||Start app
printf "\n${GREEN}Start||Restart app ${NC}\n"
service sabz.city restart || true
