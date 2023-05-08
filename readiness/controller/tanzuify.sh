kubectl apply -f https://github.com/vmware-tanzu/carvel-kapp-controller/releases/latest/download/release.yml
kubectl apply -f https://github.com/carvel-dev/secretgen-controller/releases/latest/download/release.yml

imgpkg pull -b devtester.azurecr.io/readiness:dev.0.0_1 -o readiness-pkg
ytt -f readiness-pkg/config/ | kbld -f - -f readiness-pkg/.imgpkg/images.yml | kubectl apply -f-