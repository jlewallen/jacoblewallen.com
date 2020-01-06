#!/bin/bash

set -xe

pushd ..

source private.env

for pf in `ls jacoblewallen.com/public/albums/*.html`; do
	echo $pf

	build/secure --passphrase ${PASSPHRASE} --plaintext $pf --ciphertext $pf.aes

	mv $pf.aes $pf
done

popd
