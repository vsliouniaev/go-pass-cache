window.onload = function () {
    let naclfac = nacl_factory.instantiate(decrypt)

    const onSubmit = async (e) => {
        e.preventDefault()
        let nacl = await naclfac
        let id = nacl.to_hex(nacl.random_bytes(32))
        let key = nacl.random_bytes(32)

        let message = nacl.encode_utf8(document.getElementById('data').value)
        let nonce = nacl.crypto_secretbox_random_nonce()
        let cyphertext = nacl.crypto_secretbox(message, nonce, key)

        let xhttp = new XMLHttpRequest()
        xhttp.open("POST", "/")
        xhttp.setRequestHeader("Content-type", "application/json")
        xhttp.send(JSON.stringify({id: id, data: nacl.to_hex(cyphertext) + " " + nacl.to_hex(nonce)}))
        document.getElementById('result').removeAttribute("hidden")
        document.getElementById('accessUrl').innerHTML =
            window.location.origin + '?' + encodeURIComponent(id) + '#' + nacl.to_hex(key)
        document.getElementById('inputs').innerHTML = ""

        copyToClipboard()
    }

    const onShowQR = async (e) => {
        let btn = document.getElementById("qrButton")
        btn.setAttribute("hidden", "hidden")
        let qrCanvas = document.getElementById('qrCanvas')
        qrCanvas.parentElement.removeAttribute("hidden")
        new QRious({element: qrCanvas, value: getURL(), size: 200});
    }

    // Attach button click handler to submit action
    const form = document.getElementById("form")
    if (form !== null) {
        form.addEventListener("submit", onSubmit)
    }

    // Attach Shift + Return handler to submit action
    const data = document.getElementById("data")
    if (data !== null && form !== null) {
        data.focus()
        data.addEventListener("keypress", async (e) => {
            if (e.keyCode === 13 && e.shiftKey) {
                await onSubmit(e)
            }
        })
    }

    // Attach QR code
    const qrButton = document.getElementById("qrButton")
    if (qrButton != null) {
        qrButton.addEventListener("click", onShowQR)
    }

    // Decrypt if data is present
    function decrypt(nacl) {
        let pass = window.location.toString().split('#')
        if (pass.length === 2) {
            let e = document.getElementById('encrypted')
            if (e !== null) {
                let s = e.value.split(" ")
                document.getElementById('data').value =
                    nacl.decode_utf8(
                        nacl.crypto_secretbox_open(
                            nacl.from_hex(s[0]),
                            nacl.from_hex(s[1]),
                            nacl.from_hex(pass[1])))
            }
        }
    }

    function copyToClipboard() {
        if (!navigator.clipboard) {
            let aux = document.createElement("input")
            aux.setAttribute("value", getURL())
            document.body.appendChild(aux)
            aux.select()
            document.execCommand("copy")
            document.body.removeChild(aux)
        } else {
            navigator.clipboard.writeText(getURL()).then(
                function () {
                    // OK
                })
                .catch(
                    function () {
                        // TODO: Print to screen instead?
                    });
        }


    }

    function getURL() {
        return document.getElementById('accessUrl').innerHTML
    }
}
