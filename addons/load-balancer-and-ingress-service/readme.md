# TKG Addons for AKO

## How to install and use the AKO addon using ytt

- Create the ytt value yaml(`values.yaml`) based on the example from [here](./examples), and substitute the values for your case
- Run `ytt --ignore-unknown-comments -f ../load-balancer-and-ingress-service/templates -f ../load-balancer-and-ingress-service/examples/values.yaml > ako.yaml`
- Run `kubectl apply -f ako.yaml`. Once you have applied the `ako.yaml` you should be able to see required resources are creating. 