# TKG Addons for AKO-Operator

## How to install and use the AKO-Operator addon using ytt

- Create the ytt value yaml(`values.yaml`) based on the example from [here](./examples), and substitute the values for your case
- Run `ytt --ignore-unknown-comments -f ../ako-operator/templates -f ../ako-operator/examples/values.yaml > ako-operator.yaml`
- Run `kubectl apply -f ako-operator.yaml`. Once you have applied the `ako-operator.yaml` you should be able to see required resources are creating. 