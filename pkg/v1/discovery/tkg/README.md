# TKG Discovery

This repository is a collection of constants and functions which expose Discovery queries aimed at making it simple to extend and integrate with TKG.


Initialize a new new TKGDiscovery client with a context, a dynamic client and a discovery client. Then, you can run any existing queries:
```
tkg := NewTKGDiscovery(ctx, dynamicClient, discoveryClient)

if !tkg.IsProviderFoo() {
	log.Fatal("Provider Type is Foo")
}
```
