# How to update tanzu core to a new version or commit in a TKG Release?

## Steps

1. checkout bolt-cli repo, `git clone git@gitlab.eng.vmware.com:TKG/bolt/bolt-cli.git`
2. compile bolt-cli: `make build`
3. sync bolt-release-yamls `hack/sync-release-yaml.sh`
4. run cmd to generate the configure files
    1. If you want to build a commit/tag/branch in the `https://github.com/vmware-tanzu/tanzu-framework` repo, please use  
    ```bash
    ./bin/bolt tkgbuild --releaseYaml=tkg-v1.5.0-zshippable.yaml \
    --tanzuFrameworkUpstreamVersion=0d6405000fe81c7d283dcbe45aa2ca1f6be50ee0 \
    --stagingImageRepo=projects-stg.registry.vmware.com/tkg \
    --bomImageRepo=projects-stg.registry.vmware.com/tkg --updateVersionMap=true
    ```
    2. If you want to build a commit from your PR, and your PR is based on your forked tanzu-framework repo, bolt-cli requires github username and github token (read access is enough) to query the github API to get your PR's infomation (for example, whats the source and target repo url). You need to create your own github token using your own github account. Once you have the token, set these 2 env vars first then run the `tkgbuild` cmd
    ```bash
    export GITHUB_USERNAME=your_github_username
    export GITHUB_TOKEN=your_github_token  
    ./bin/bolt tkgbuild --releaseYaml=tkg-v1.5.0-zshippable.yaml \
    --tanzuFrameworkUpstreamVersion=8a00bf51387d284a94f9ba086106fc7260f47a31 \
    --tanzuFrameworkPullRequest=526 \
    --stagingImageRepo=projects-stg.registry.vmware.com/tkg \
    --bomImageRepo=projects-stg.registry.vmware.com/tkg --updateVersionMap=true
    ```
5. open MR with newly generated configure files in yaml folder.
