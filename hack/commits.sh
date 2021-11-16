#!/bin/bash
for i in `git log | grep commit | head -n 100 | cut -d' ' -f 2 ` ; do

	for j in "ytt" "test" "packag" "app" ; do
		if `git show --stat $i --raw | grep -q $j` ; then
			echo "$j -> c_$i ;"
		fi
	done

done

