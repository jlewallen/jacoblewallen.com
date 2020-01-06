#!/bin/bash

set -xe

pushd ..

source private.env

for pf in `find jacoblewallen.com/public/albums/ -iname "*.html"`; do
	echo $pf

	build/secure --passphrase ${PASSPHRASE} --plaintext $pf --ciphertext $pf.aes

	mv $pf.aes $pf
done

popd
