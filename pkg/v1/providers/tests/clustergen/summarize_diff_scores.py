#!/usr/bin/env python3

# Copyright 2022 Tanzu Framework Contributors.
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

import csv
import os
import sys
import hashlib

infras = ["AWS", "AZURE", "VSPHERE"]
val_types = ["MIN", "AVG", "MAX"]
scores = {"AWS": [0,0,0],
          "AZURE": [0,0,0],
          "VSPHERE": [0,0,0]}

def summarize_diff_scores(test_dir, output_file):
    test_num = 0
    for infra in infras:
        csv_file = '{}/{}_diff_summary.csv'.format(test_dir, infra)
        same_percentages = []
        try:
            with open(csv_file, 'r') as file:
                my_reader = csv.DictReader(file, delimiter=',')
                for var_dict in my_reader:
                    cfg_args = []
                    for k, v in sorted(var_dict.items()):
                        same_percentages.append(int(var_dict["SAME"]))
                        if len(same_percentages) > 0:
                            scores[infra] = [min(same_percentages),
                                             int(sum(same_percentages)/len(same_percentages)),
                                             max(same_percentages)]
        except:
            print("Error processing {}, skipped".format(csv_file))
            pass

    print("scores = {}".format(scores))
    headers = []
    vals = []
    for infra in infras:
        for t in val_types:
            headers.append("{}_{}".format(infra, t))
        vals.extend(scores[infra])

    with open(output_file, "w+") as w:
        w.write("{}\n{}\n".format(",".join(headers), ",".join([str(x) for x in vals])))

def main():
    if len(sys.argv) != 3:
        print("Usage {} test_data_dir output_file_path".format(sys.argv[0]))
        sys.exit(1)
    else:
        summarize_diff_scores(sys.argv[1], sys.argv[2])

if __name__ == "__main__":
    main()
