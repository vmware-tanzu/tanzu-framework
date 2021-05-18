#!/bin/bash

#### Setup environmental variables
#GLOBAL_ENVS="../../../globals.env";
#if [[ -f ${GLOBAL_ENVS} ]]
#then
    # shellcheck source=/dev/null       
#    source ${GLOBAL_ENVS}
#fi

# Load variables for e2e tests
# Check if VTAAS_USER is defined
source ./e2e_config

IF_VTAAS=$(echo "${E2E_SPEC}" | grep 'vtaas')
if [ "x${IF_VTAAS}" != "x" ]
then
    if [ "x${VTAAS_USER}" == "x" ]
    then
        echo "Please export VTAAS_USER(your email) in cmdline or e2e_config"
        exit 1
    else
        echo "VTAAS_USER is ${VTAAS_USER}"
    fi
fi

#### Passing all OS env variables to protractor
# node ./src/utils/os-env-dumper.js

# Load variables for e2e tests
source ./e2e_config

# Check if VTAAS_USER is defined
IF_VTAAS=$(echo "${E2E_SPEC}" | grep 'vtaas')
if [ "x${IF_VTAAS}" != "x" ]
then
    if [ "x${VTAAS_USER}" == "x" ]
    then
        echo "Please export VTAAS_USER(your email) in cmdline or e2e_config"
        exit 1
    else
        echo "VTAAS_USER is ${VTAAS_USER}"
    fi
fi

#### Clean up existing webdriver instance if any
pgrep webdriver | xargs kill  > /dev/null 2<&1

echo "Update webdriver-manager..."
webdriver-manager update
#### Start selenium server

echo "Start selenium server..."
(webdriver-manager start &)
while ! nc -z localhost 4444; do sleep 1; done

##### Executing e2e testing
echo "Executing e2e testing..."
npx protractor ./protractor.conf.js

#### Clean up testing fixtures
webdriver-manager clean  > /dev/null 2<&1
webdriver-manager shutdown  > /dev/null 2<&1

echo "Clean up testing fixtures..."
lsof -i tcp:4444 | grep LISTEN | awk '{print $2}' | xargs kill  > /dev/null 2<&1

#### Done
echo "e2e testing done!!"
