#!/usr/bin/env python3

# Copyright 2020 The TKG Contributors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


# In TKG 1.2, we provided a way for users to customize the Azure VM images by editing the BOM.

# This script is used to duplicate BOM files for a given k8s version which are populated with various test configurations
# which further can be used in clustergen tests.

import os
import yaml
import sys

def get_bom_dir():
    return os.path.join(sys.argv[1], "bom")

def get_default_tkr_bom():
    bomDir = get_bom_dir()
    for bomFile in os.listdir(bomDir):
        if(bomFile.startswith("tkr-bom")):
            return bomFile
    return None

def main():
    bomFile = get_default_tkr_bom()
    bomDir = get_bom_dir()
    bomFilePath = os.path.join(bomDir, bomFile)
    with open(bomFilePath) as file:
        bom = yaml.safe_load(file)
    osInfo = bom["azure"][0]["osinfo"]

    # Create a file with no azure VM image info in BOM
    del bom["azure"]
    bom["release"]["version"] = "v0.0.0+no-azure-image"
    with open(os.path.join(bomDir, "bom-clustergen-no-azure-image-test.yaml"), 'w') as file:
        documents = yaml.dump(bom, file)

    # Create a file with marketplace azure VM image info in BOM
    bom["azure"] = [{
        "publisher": "bom-image-publisher",
        "offer":"bom-image-offer",
        "sku":"bom-image-sku",
        "version":"bom-image-version",
        "thirdPartyImage":True,
        "osinfo": osInfo
    }]
    bom["release"]["version"] = "v0.0.0+marketplace-image"
    with open(os.path.join(bomDir, "bom-clustergen-marketplace-image-test.yaml"), 'w') as file:
        documents = yaml.dump(bom, file)

    # Create a file with marketplace azure VM image(no thirdPartyImage flag) info in BOM
    bom["azure"] = [{
        "publisher": "bom-image-publisher",
        "offer":"bom-image-offer",
        "sku":"bom-image-sku",
        "version":"bom-image-version",
        "osinfo": osInfo
    }]
    bom["release"]["version"] = "v0.0.0+marketplace-image-no-thirdpartyimage"
    with open(os.path.join(bomDir, "bom-clustergen-marketplace-no-thirdparty-image-test.yaml"), 'w') as file:
        documents = yaml.dump(bom, file)

    # Create a file with shared gallery azure VM image info in BOM
    bom["azure"] = [{
        "resourceGroup": "bom-resource-group",
        "name":"bom-image-name",
        "subscriptionID":"bom-subscription-id",
        "version":"bom-image-version",
        "gallery":"bom-image-gallery",
        "osinfo": osInfo
    }]
    bom["release"]["version"] = "v0.0.0+shared-gallery-image"
    with open(os.path.join(bomDir, "bom-clustergen-shared-gallery-image-test.yaml"), 'w') as file:
        documents = yaml.dump(bom, file)

if __name__ == "__main__":
    main()
