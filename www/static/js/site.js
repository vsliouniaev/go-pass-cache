// /Get handler

window.onload = function () {
    const onSubmit = async (e) => {
        e.preventDefault()
        let nacl = await nacl_factory.instantiate(function () {
        })

        let id = nacl.to_hex(nacl.random_bytes(32))
        let key = nacl.random_bytes(32)

        let message = nacl.encode_utf8(document.getElementById('data').value)
        let nonce = nacl.crypto_secretbox_random_nonce()
        let cyphertext = nacl.crypto_secretbox(message, nonce, key)

        let xhttp = new XMLHttpRequest()
        xhttp.open("POST", "/", false)
        xhttp.setRequestHeader("Content-type", "application/json")
        xhttp.send(JSON.stringify({id: id, data: nacl.to_hex(cyphertext) + " " + nacl.to_hex(nonce)}))
        document.getElementById('result').removeAttribute("hidden")
        document.getElementById('accessUrl').innerHTML =
            window.location.origin + '?' + encodeURIComponent(id) + '#' + nacl.to_hex(key)
        document.getElementById('inputs').innerHTML = ""
    }

    // Set handling
    const form = document.getElementById("form")
    if (form !== null) {
        form.addEventListener("submit", onSubmit)
    }

    const data = document.getElementById("data")
    if (data !== null && form !== null) {
        data.focus()
        data.addEventListener("keypress", async (e) => {
            if (e.keyCode === 13 && e.shiftKey) {
                await onSubmit(e)
            }
        })
    }

    let pass = window.location.toString().split('#')
    if (pass.length === 2) {
        let e = document.getElementById('encrypted')
        if (e !== null) {
            let e = e.value.split(" ")
            nacl_factory.instantiate(function (nacl) {
                document.getElementById('data').value =
                    nacl.decode_utf8(
                        nacl.crypto_secretbox_open(
                            nacl.from_hex(e[0]),
                            nacl.from_hex(e[1]),
                            nacl.from_hex(pass[1])))
            })
        }
    }
}

function copyToClipboard() {
    let aux = document.createElement("input")
    aux.setAttribute("value", document.getElementById('accessUrl').innerHTML)
    document.body.appendChild(aux)
    aux.select()
    document.execCommand("copy")
    document.body.removeChild(aux)

    let button = document.getElementById('copy-button')
    button.setAttribute('disabled', 'disabled')
    button.innerHTML = 'Copied to clipboard'
}