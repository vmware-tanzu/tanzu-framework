#!/bin/bash

wide=false
if [[ $2 ]]; then
    wide=true
fi

print_readiness_providers () {
    if [ $wide == true ]; then
        kubectl get readinessproviders -o json > readinessproviders.json
        echo -e "NAME\t\t\tSTATE\t\tPASSING\tCONDITIONS"
        for row in $(cat readinessproviders.json | jq -r '.items[] | @base64'); do
            _parse_readinessproviders() {
                total_conditions=$(echo ${row} | base64 --decode | jq -r '.spec.conditions | length')
                conditions_passing=$(echo ${row} | base64 --decode | jq -r '.status.conditions | select(.) | map(select(.state == "success")) | length')
                name=$(echo ${row} | base64 --decode | jq -r '.metadata.name')
                conditions=$(echo ${row} | base64 --decode | jq -r -c '[.spec.conditions[].name]')
                state=$(echo ${row} | base64 --decode | jq -r '.status.state')
                printf "%-20s %10s %11s     %s\n" $name $state "${conditions_passing}/${total_conditions}" $conditions
            }
            _parse_readinessproviders
        done
        rm -rf readinessproviders.json
   else
        kubectl get readinessproviders;
   fi
}

print_readiness () {
    if [ $wide == true ]; then
        kubectl get readiness -o json > readiness.json
        echo -e "NAME\t\tREADY\tACTIVE\tCHECKS"
        for row in $(cat readiness.json | jq -r '.items[] | @base64'); do
            _parse_readiness() {
                total_checks=$(echo ${row} | base64 --decode | jq -r '.spec.checks | length')
                checks_passing=$(echo ${row} | base64 --decode | jq -r '.status.checkStatus | select(.) | map(select(.status == true)) | length')
                name=$(echo ${row} | base64 --decode | jq -r '.metadata.name')
                checks=$(echo ${row} | base64 --decode | jq -r -c '[.spec.checks[].name]')
                ready=$(echo ${row} | base64 --decode | jq -r '.status.ready')
                printf "%-10s %10s %5s     %s\n" $name $ready "${checks_passing}/${total_checks}" $checks
            }
            _parse_readiness
        done
        rm -rf readiness.json
   else
        kubectl get readiness;
   fi
}

if [ $1 == "readiness" ]; then
    print_readiness; 
else 
    print_readiness_providers;
fi