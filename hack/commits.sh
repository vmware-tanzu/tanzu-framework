#!/bin/bash
for i in `git log | grep commit | head -n 100 | cut -d' ' -f 2 ` ; do

	if `git show --stat $i --raw | grep -q addons` ; then
		echo -n "addons_"
	fi
	if `git show --stat $i --raw | grep "|" | grep -q providers` ; then
		echo -n "providers_"
	fi
	if `git show --stat $i --raw | grep "|" | grep -q pkg/v1/tkg/web`; then
		echo -n "web_"
	fi
	if `git show --stat $i --raw | grep "|" | grep -q cli`; then 
		echo -n "client_"
	fi
	if `git show --stat $i --raw | grep "|" | grep -q crd`; then
		echo -n "api_crd_"
	fi
	if `git show --stat $i --raw | grep "|" | grep -q test` ; then 
		echo -n "tests_"
	fi
	if `git show --stat $i --raw | grep "|" | grep -q pkg/v1/tkg` ; then
		echo -n "pkg_v1_tkg_"
	fi
	if `git show --stat $i --raw | grep "|" | grep -q hack` ; then
		echo -n "hack_test_stuff_"
	fi
	if `git show --stat $i --raw | grep -q "md"` ; then 
		echo -n "docs_"
	fi
	if `git show --stat $i --raw | grep "|" | grep -q packages` ; then
		echo -n "packages_"
	fi
	if `git show --stat $i --raw | grep -q ytt` ; then
		echo -n "ytt_" 
	fi
	if `git show --stat $i --raw | grep -q web`; then
		echo -n "webstuff_"
	fi
	echo " -> c_$i ;"

	for j in "ytt" "test" "packag" "app" ; do
		if `git show --stat $i --raw | grep -q $j` ; then
			echo "$j -> c_$i ;"
		fi
	done

done

