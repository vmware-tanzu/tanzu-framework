# TKG Addons for Metrics Server

## How to install and use the Metrics-Server addon using ytt

- Create the ytt value yaml(`values.yaml`) based on the example from [here](./examples), and substitute the values for your case
- Run `ytt --ignore-unknown-comments -f ../ytt-common-libs -f ../metrics-server/templates -f values.yaml > metrics-server.yaml`
- Run `kubectl apply -f metrics-server.yaml`. Once you have applied the `metrics-server.yaml` you should be able to see required resources are creating. 
- User `kubectl top node` and `kubectl top pods -A` to get the memory and CPU usage info.