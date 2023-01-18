
## Release graph

This directory holds all the release configurations used by Bolt.  
A release is a collection of components that make up a _release graph_. A release graph is also referred to as a _build graph_ or a _dependency graph_.  
A release graph is a Directed Acyclic Graph that represents the build dependencies between all the components involved.  

## Example Release Graph  

<img src="example_release_graph.png" alt="Example Release Graph" width="300" height="350"/>  

The components in a release graph are processed in the topological sort order. Any components that can be processed concurrently will be processed concurrently (upto a maximum of `workers`, defaults to 5).  

For more details on how each component is processed please refer to _how a component is built_ sections [here](../component/README.md).

## Create Your Own Graph

Other than propose change to existing release graph, Bolt support every developer to create their own graph by openning merge requst, it started with creating a new file under the release folder, this new yaml file will refer one or more new root component + version combination.

one typical use case of creating new graph is: create a new TKr, and linked to existing graph or create a new graph.

bolt-cli has a command call `tkrbuild` to automatically generate new TKr merge request.

