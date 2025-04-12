let elementVisible = null
let aCallOnHide = null

function showPopUpById(id, callOnHide) {
    hidePopUp()
    setTimeout(function () {
        elementVisible = document.getElementById(id);
        if (elementVisible != null) {
            elementVisible.style.visibility = "visible"
            aCallOnHide = callOnHide
        }
    })
}

document.addEventListener("click", (evt) => {
    if (elementVisible != null) {
        let targetEl = evt.target; // clicked element
        do {
            if (targetEl === elementVisible) {
                // This is a click inside, does nothing, just return.
                return;
            }
            // Go up the DOM
            targetEl = targetEl.parentNode;
        } while (targetEl);
        // This is a click outside.
        hidePopUp()
        evt.preventDefault()
    }
});

function hidePopUp() {
    if (elementVisible != null) {
        elementVisible.style.visibility = "hidden"
        elementVisible = null
        if (aCallOnHide != null) {
            aCallOnHide();
            aCallOnHide = null
        }
    }
}

function loadExample(name) {
    hidePopUp();
    let source = document.getElementById('source');
    fetchHelper("/example/", name, function(code) {
        source.value = code;
        runSource();
    });
}

function runSource() {
    let source = document.getElementById('source');
    let result = document.getElementById('result');
    fetchHelper("/execute/", source.value, a => result.innerHTML = a);
}

function fetchHelper(url, data, target) {
    let formData = new FormData();
    formData.append('data', data);

    fetch(url, {body: formData, method: "post", signal: AbortSignal.timeout(3000)})
        .then(function (response) {
            if (response.status !== 200) {
                window.location.reload();
                return;
            }
            return response.text();
        })
        .catch(function (error) {
            window.location.reload();
        })
        .then(function (html) {
            target(html);
        })
}
