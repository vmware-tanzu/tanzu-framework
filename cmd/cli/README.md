# Tanzu CLI

The [Tanzu CLI](https://docs.google.com/document/d/1X34ZNkPG_kEMSySpFjAQsmX2Xn1dXTksbVxXUgUk-QM/edit?usp=sharing)

## Installation
### MacOS
```shell
curl -o tanzu https://storage.googleapis.com/tanzu-cli/artifacts/core/latest/tanzu-core-darwin_amd64 && \
    mv tanzu /usr/local/bin/tanzu && \
    chmod +x /usr/local/bin/tanzu
```
### Linux
#### i386
```shell
curl -o tanzu https://storage.googleapis.com/tanzu-cli/artifacts/core/latest/tanzu-core-linux_386 && \
    mv tanzu /usr/local/bin/tanzu && \
    chmod +x /usr/local/bin/tanzu
```
#### AMD64
```shell
curl -o tanzu https://storage.googleapis.com/tanzu-cli/artifacts/core/latest/tanzu-core-linux_amd64 && \
    mv tanzu /usr/local/bin/tanzu && \
    chmod +x /usr/local/bin/tanzu
```

### Windows
Windows executable can be found at https://storage.googleapis.com/tanzu-cli/artifacts/core/latest/tanzu-core-windows_amd64.exe


## Contribution
See the [Tanzu CLI Implementation Guide](/docs/cli/plugin_implementation_guide.md)
