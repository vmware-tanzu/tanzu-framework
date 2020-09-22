# ad hoc meeting notes

// add go get to main repository
// add build/test directory
// add more tests for everthing
// Owners files subdirectories
// Can we grab pieces of the kubernetes bots? pull-approve
// Top level readme
// Docs release team // gitbook mdbook // steuart clemet IX
// goreleaser for changelogs
// weekly code walkthrough -- maybe here maybe other meeting
// remove `.cloud` from api group -- will pursue .tanzu with joe
// where to store the images -- dockerhub? we have org license -- habor? -- 
// commit messaging tool -- gh has recommendations and templating -- PR templates
    -- cloudprovider/vsphere has good tooling
    -- CAPI issue templates
// examples directory
// define standard errors and logging
    -- wrapping logging is challenging
    -- can standardize formatting and fix logging upstream
    -- figure out a pattern for mingling with the upstream logs
    -- pkg/errors for errors
// big kubeconfig with multiple contexts, and then kubctl execute to get session token
// factor out session management
// how does TAS integrate
    -- just path auth to the next cli
    -- do we proxy the commands to it
    -- do we reimplement all the commands


// TMC feedback

* May have different developer/operator 
* How do we handle auth?
* Leading or follower? both
* provide a detailed overview
* update strategy
* who does the docs
* 