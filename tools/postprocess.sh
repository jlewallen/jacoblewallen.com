#!/bin/bash

set -xe

pushd ..

source private.env

for pf in `find jacoblewallen.com/public/albums/ -iname "*.html"`; do
	echo $pf

	build/secure --inline --passphrase ${PASSPHRASE} --plaintext $pf --ciphertext $pf.aes

	if [ -f $pf.aes ]; then
		mv $pf.aes $pf
	fi
done

popd
