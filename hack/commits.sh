for i in `git log | grep commit | head -n 50 | cut -d' ' -f 2 ` ; do

	if `git show --stat $i --raw | grep -q addons` ; then
		echo "addons $i"
	fi
	if `git show --stat $i --raw | grep "|" | grep -q providers` ; then
		echo  "providers $i"
	fi
	if `git show --stat $i --raw | grep "|" | grep -q pkg/v1/tkg/web`; then
		echo "web $i"
	fi
	if `git show --stat $i --raw | grep "|" | grep -q cli`; then 
		echo "client $i"
	fi
	if `git show --stat $i --raw | grep "|" | grep -q crd`; then
		echo "api/crd $i"
	fi
	if `git show --stat $i --raw | grep "|" | grep -q test` ; then 
		echo "tests $i"
	fi
	if `git show --stat $i --raw | grep "|" | grep -q pkg/v1/tkg` ; then
		echo "pkg/v1/tkg/"
	fi
	if `git show --stat $i --raw | grep "|" | grep -q hack` ; then
		echo "hack/test stuff $i"
	fi
	if `git show --stat $i --raw | grep -q "md"` ; then 
		echo "docs $i"
	fi
	if `git show --stat $i --raw | grep "|" | grep -q packages` ; then
		echo "packages $i"
	fi
	if `git show --stat $i --raw | grep -q ytt` ; then
		echo "ytt $i" 
	fi
	if `git show --stat $i --raw | grep -q web`; then
		echo "web stuff $i"
	fi
done

