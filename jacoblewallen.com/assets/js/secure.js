class SecureBlock {
	constructor(el, encrypted) {
		this.el = el;
		this.encrypted = encrypted;
	}

	unlock() {
		const session = this.loadSession();

		if (session) {
			if (this.reveal(session.passphrase)) {
				return true;
			}
		}

		const form = this.el.find(".encrypted-form");
		const self = this;

		form.show();

		document.getElementById('encrypted-form').addEventListener('submit', function(e) {
			e.preventDefault();

			const passphrase = document.getElementById('encrypted-password').value;

			if (self.reveal(passphrase)) {
				self.saveSession({
					passphrase: passphrase
				});
			}
		});

		return false;
	}

	reveal(passphrase) {
		const decrypted = this.verifyAndDecrypt(this.encrypted, passphrase);
		if (decrypted) {
			this.el.replaceWith($(decrypted))
			return true;
		}

		return false;
	}

    decrypt(encrypted, passphrase) {
        const salt = CryptoJS.enc.Hex.parse(encrypted.substr(0, 32));
        const iv = CryptoJS.enc.Hex.parse(encrypted.substr(32, 32))
        const ciphertext = encrypted.substring(64);

		const keySize = 256;
		const iterations = 4096;
        const key = CryptoJS.PBKDF2(passphrase, salt, {
            keySize: keySize / 32,
            iterations: iterations
        });

        const decrypted = CryptoJS.AES.decrypt(ciphertext, key, {
            iv: iv,
            padding: CryptoJS.pad.Pkcs7,
            mode: CryptoJS.mode.CBC
        });

		const utf8 = decrypted.toString(CryptoJS.enc.Utf8);

        return utf8;
    }

	verifyAndDecrypt(encrypted, passphrase) {
		const encryptedHMAC = encrypted.substring(0, 64);
		const encryptedHTML = encrypted.substring(64);
		const passphraseHash = CryptoJS.SHA256(passphrase).toString();
		const decryptedHMAC = CryptoJS.HmacSHA256(encryptedHTML, passphraseHash).toString();

		if (decryptedHMAC !== encryptedHMAC) {
			return null;
        }

		return this.decrypt(encryptedHTML, passphrase);
	}

	loadSession() {
		const raw = window.localStorage.getItem('session');
		if (!raw) {
			return null;
		}

		return JSON.parse(raw);
	}

	saveSession(session) {
		window.localStorage.setItem('session', JSON.stringify(session));

		console.log("saved session", session);

		return true;
	}
}
