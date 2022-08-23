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

import csv
import os
import sys
import hashlib


def write_test_cases(params_file, test_dir):
    test_num = 0
    with open(params_file, 'r') as file:
        my_reader = csv.DictReader(file, delimiter=',')
        for var_dict in my_reader:
            test_num+=1
            cmd_args = []
            cfg_args = []
            dict_contents = ""
            for k, v in sorted(var_dict.items()):
                dict_contents += '{}{}'.format(k, v)
                if k == "_CNAME":
                    cmd_args.insert(0, v)
                elif k == "_PLAN":
                    cmd_args.append('{} {}'.format("--plan", v))
                elif k == "_INFRA":
                    cmd_args.append('{} {}'.format("-i", v))

                if k.startswith('--'):
                    if v != "NOTPROVIDED":
                        k = k.lower()
                        if k.startswith('--enable-'):
                            cmd_args.append('{}={}'.format(k, v))
                        else:
                            cmd_args.append('{} {}'.format(k, v))
                else:
                    # hack to workaround the problem with pict where there is
                    # no escape char for comma and double quote.
                    # sets "AZURE_CUSTOM_TAGS" to "tagKey1=tagValue1, tagKey2=tagValue2"

                    if k == "AZURE_CUSTOM_TAGS" and v.startswith('tagKey1='):
                        cfg_args.append('{}: {}'.format(k, 'tagKey1=tagValue1, tagKey2=tagValue2'))
                    elif v != "NA":
                        cfg_args.append('{}: {}'.format(k, v.replace("<comma>", ",").replace("<qq>", "\"")))

            testid = int(hashlib.sha256(dict_contents.encode('utf-8')).hexdigest(), 16) % 10**8
            filename = "%8.8d.case" % (testid)
            with open(os.path.join(test_dir, filename), "w") as w:
                w.write('#! ({}) EXE: {}\n\n'.format("%4.4d" % test_num, " ".join(cmd_args)))
                w.write('{}\n'.format("\n".join(cfg_args)))

def main():
    if len(sys.argv) != 3:
        print("Usage {} csv_params_file test_data_dir".format(sys.argv[0]))
        sys.exit(1)
    else:
        write_test_cases(sys.argv[1], sys.argv[2])

if __name__ == "__main__":
    main()
