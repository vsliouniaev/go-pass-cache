// /Get handler

window.onload = function () {
    var encrypted = document.getElementById('encrypted').value;
    var pass = window.location.toString().split('#');

    nacl_factory.instantiate(function (nacl) {
        k = nacl.random_bytes(32);
        m = nacl.encode_utf8("message");
        n = nacl.crypto_secretbox_random_nonce();
        c = nacl.crypto_secretbox(m, n, k);

        box = nacl.to_hex(c)
        non = nacl.to_hex(n)
        pas = nacl.to_hex(k)

        dbox = nacl.from_hex(box)
        dnon = nacl.from_hex(non)
        dpas = nacl.from_hex(pas)
        m1 = nacl.crypto_secretbox_open(dbox, dnon, dpas);
        // "message" === nacl.decode_utf8(m1); // always true
        alert(nacl.decode_utf8(m1));
    });

    if (pass.length === 2) {
        document.getElementById('data').value = sjcl.decrypt(pass[1], encrypted);
    }
}

// /Set handler
var showCreds = false;
var stop = false;
var urlInteverval = (Math.random() * 10) + 5;
var passInterval = (Math.random() * 10) + 5;
var pass = '';
var raw = '';
var fullUrl = '';
var urlP = Uheprng();
var passwordP = Uheprng();

// ---------------------------------------------------------------------
// Set up pseudo random number generators for the url token and password
// ---------------------------------------------------------------------

urlP.initState();
urlP.addEntropy(getFromBrowserIfAvailable());
urlP.addEntropy(urlEntropy);

passwordP.initState();
passwordP.addEntropy(getFromBrowserIfAvailable());
passwordP.addEntropy(passEntropy);

function urlG() {
    urlP.addEntropy();
    if (!stop) {
        setTimeout(urlG, urlInteverval);
    }
}

function passG() {
    passwordP.addEntropy();
    if (!stop) {
        setTimeout(passG, passInterval);
    }
}

function getFromBrowserIfAvailable() {
    var cryptoObj = window.crypto || window.msCrypto;
    if (!cryptoObj) {
        return Math.random();
    }

    var array = new Uint32Array(64);
    window.crypto.getRandomValues(array);
    return array;
}

urlG();
passG();

// ---------------------------------------------------------------------
//                        Interaction handlers
// ---------------------------------------------------------------------

function beforeFormSubmit() {
    // stopPrngs();
    nacl_factory.instantiate(function (nacl) {


        raw = sjcl.codec.base64.fromBits(sjcl.hash.sha256.hash(urlP.string(64)), true);
        var id = encodeURIComponent(nacl.random_bytes(32));
        fullUrl = url + '?id=' + id;
        pass = sjcl.codec.base64.fromBits(sjcl.hash.sha256.hash(passwordP.string(64)), true);
        fullUrl += '#' + pass;

        var data = document.getElementById('data').value;
        document.getElementById('id').value = raw;


        let k = nacl.random_bytes(32);
        let m = nacl.encode_utf8(data);
        let n = nacl.crypto_secretbox_random_nonce();
        let c = nacl.crypto_secretbox(m, n, k);


        box = nacl.to_hex(c)
        non = nacl.to_hex(n)
        pas = nacl.to_hex(k)

        document.getElementById('encrypted').value = {box:box,non:non}
        showCreds = true;
    })

    // document.getElementById('encrypted').value = sjcl.encrypt(pass, data);
    // showCreds = true;
}

// function stopPrngs() {
//     stop = true;
//     raw = sjcl.codec.base64.fromBits(sjcl.hash.sha256.hash(urlP.string(64)), true);
//     var id = encodeURIComponent(raw);
//     fullUrl = url + '?id=' + id;
//     pass = sjcl.codec.base64.fromBits(sjcl.hash.sha256.hash(passwordP.string(64)), true);
//     fullUrl += '#' + pass;
// }

function afterFormSubmit() {
    if (showCreds) {
        document.title = "passcache";
        document.getElementById('result').removeAttribute("hidden");
        document.getElementById('accessUrl').innerHTML = fullUrl;
        document.getElementById('inputs').innerHTML = "";
    }
    showCreds = false;
}

function copyToClipboard() {

    var aux = document.createElement("input");
    aux.setAttribute("value", document.getElementById('accessUrl').innerHTML);
    document.body.appendChild(aux);
    aux.select();
    document.execCommand("copy");
    document.body.removeChild(aux);

    var button = document.getElementById('copy-button');
    button.setAttribute('disabled', 'disabled');
    button.innerHTML = 'Copied to clipboard';
}