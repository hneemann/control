function updateRow() {
    let textareaElement = document.getElementById("source");
    let rowElement = document.getElementById("row");
    let line = 1;
    let selectionStart = textareaElement.selectionStart;
    let text = textareaElement.value;
    for (let i = 0; i < selectionStart; i++) {
        if (text[i] === '\n') {
            line++;
        }
    }
    rowElement.textContent = line.toString();
}

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

// This function is used to set the name of the file in the save dialog.
// It is called when the user clicks on a file in the list of files.
function setName(name) {
    let filename = document.getElementById('saveDialogFilename');
    filename.value = name;
}

let loadedCode = "";
let loadedName = "";

function cleanString(str) {
    return str.replace(/\r\n/g, "\n").trim();
}

function setSource(name, code) {
    let source = document.getElementById('source');
    source.value = code;
    loadedCode = cleanString(code);
    loadedName = name;
    let label = document.getElementById('filenameLabel');
    label.innerHTML = name;
    runSource();
}

let overwriteAction = null;

function checkOverwrite(action) {
    let source = document.getElementById('source');
    let srcCleaned = cleanString(source.value);
    if (loadedCode !== srcCleaned) {
        overwriteAction = action;
        showPopUpById('overwriteConfirm')
        return;
    }
    action();
}

function overwriteConfirmed() {
    hidePopUp();
    if (overwriteAction != null) {
        overwriteAction();
        overwriteAction = null;
    }
}

function newScript() {
    checkOverwrite(() => {
        setSource("", "")
    })
}

function loadExample(name) {
    hidePopUp();
    checkOverwrite(() => {
        fetchHelper("./examples/"+name+".control", function (code) {
            setSource(name, code);
        });
    });
}

function fetchHelper(url, target) {
    fetch(url)
        .then(function (response) {
            if (response.status !== 200) {
                window.location.reload();
                return;
            }
            return response.text();
        })
        .catch(function (error) {
            showPopUpById("networkError");
            target=null
        })
        .then(function (html) {
            if (target != null) {
                target(html);
            }
        })
}
