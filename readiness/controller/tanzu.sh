#!/bin/bash

## Usage
## bash tanzu.sh readiness
## bash tanzu.sh readiness <readiness-name>
## bash tanzu.sh readinessproviders
## bash tanzu.sh readinessproviders <readiness-name>


print_readiness_providers () {
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
}

print_readiness () {
    kubectl get readiness -o json > readiness.json
    echo -e "NAME\t\tREADY\tACTIVE"
    for row in $(cat readiness.json | jq -r '.items[] | @base64'); do
        _parse_readiness() {
            total_checks=$(echo ${row} | base64 --decode | jq -r '.spec.checks | length')
            checks_passing=$(echo ${row} | base64 --decode | jq -r '.status.checkStatus | select(.) | map(select(.status == true)) | length')
            name=$(echo ${row} | base64 --decode | jq -r '.metadata.name')                
            ready=$(echo ${row} | base64 --decode | jq -r '.status.ready')
            printf "%-10s %10s %5s\n" $name $ready "${checks_passing}/${total_checks}"
        }
        _parse_readiness
    done
    rm -rf readiness.json
}

print_readiness_detail () {
    kubectl get readiness $1 -o json > readiness.json
    code=$?
    if [ $code == 1 ]; then    
        exit 1 
    fi
    cat readiness.json
    echo -e "CHECK-NAME\t\t\t\t\tPROVIDERS\t\t\t\tACTIVE-PROVIDERS"
    for row in $(cat readiness.json | jq -r '.status.checkStatus[] | @base64'); do        
        _parse_readiness() {
            check_name=$(echo ${row} | base64 --decode | jq -r '.name')            
            providers=$(echo ${row} | base64 --decode | jq -r '[.providers[].name]' | jq -r tostring)
            active=$(echo ${row} | base64 --decode | jq -r '[.providers[] | select(.isActive == true) | .name]' | jq -r tostring)
            printf "%-48s%-40s%s\n" $check_name $providers $active          
        }
        _parse_readiness
    done
    rm -rf readiness.json
}

print_readiness_providers_detail () {
    kubectl get readinessproviders $1 -o json > readiness.json
    code=$?
    if [ $code == 1 ]; then    
        exit 1 
    fi
    echo -e "CONDITION\t\t\tSTATE"
    for row in $(cat readiness.json | jq -r '.status.conditions[] | @base64'); do        
        _parse_readiness() {
            condition=$(echo ${row} | base64 --decode | jq -r '.name')            
            condition_state=$(echo ${row} | base64 --decode | jq -r '.state')                        
            #echo $message      
            printf "%-32s%-16s\n" $condition $condition_state          
        }
        _parse_readiness
    done
    rm -rf readiness.json
}

if [ $1 == "readiness" ]; then
    if [ "$2a" == "a" ]; then
        print_readiness;
    else        
        print_readiness_detail $2;
    fi
else 
    if [ "$2a" == "a" ]; then
        print_readiness_providers;
    else
        print_readiness_providers_detail $2;
    fi
fi