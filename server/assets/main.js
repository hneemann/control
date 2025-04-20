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

function cleanString(str) {
    return str.replace(/\r\n/g, "\n").trim();
}

function setSource(name, code) {
    let source = document.getElementById('source');
    source.value = code;
    loadedCode = cleanString(code);
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

function loadExample(name) {
    hidePopUp();
    checkOverwrite(() => {
        fetchHelper("/example/", name, function (code) {
            setSource(name, code);
        });
    });
}

function runSource() {
    let source = document.getElementById('source');
    let result = document.getElementById('result');

    //result.innerHTML = generateOutput(source.value);
    fetchHelper("/execute/", source.value, a => result.innerHTML = a);
}

function showLoad() {
    checkOverwrite(() => {
        let formData = new FormData();
        formData.append('cmd', "loadList");
        fetchHelperForm("/files/", formData, html => {
            let files = document.getElementById('loadFileList');
            files.innerHTML = html;
            showPopUpById('loadDiv')
        })
    })
}

function showSave() {
    let label = document.getElementById('filenameLabel');
    let filename = document.getElementById('saveDialogFilename');
    filename.value = label.innerHTML;

    let formData = new FormData();
    formData.append('cmd', "saveList");
    fetchHelperForm("/files/", formData, html => {
        let files = document.getElementById('saveFileList');
        files.innerHTML = html;
        showPopUpById('saveDiv')
    })
}

function saveSource() {
    let formData = new FormData();
    let filename = document.getElementById('saveDialogFilename');
    formData.append('cmd', "exists");
    formData.append('name', filename.value);
    fetchHelperForm("/files/", formData, text => {
        if (text.trim() !== "false") {
            let label = document.getElementById('confirmName');
            label.innerHTML = filename.value;
            showPopUpById('saveConfirm')
            return;
        }
        overwriteSource();
    })
}

function overwriteSource() {
    hidePopUp();
    let formData = new FormData();
    let filename = document.getElementById('saveDialogFilename');
    let source = document.getElementById('source');
    formData.append('cmd', "save");
    formData.append('name', filename.value);
    formData.append('src', source.value);
    fetchHelperForm("/files/", formData, html => {
        let label = document.getElementById('filenameLabel');
        label.innerHTML = filename.value;
        loadedCode = source.value;
    })
}

function loadSource(name) {
    let formData = new FormData();
    formData.append('cmd', "load");
    formData.append('name', name);
    fetchHelperForm("/files/", formData, code => {
        setSource(name, code);
        hidePopUp();
    })
}

function deleteFileConfirm() {
    let filename = document.getElementById('saveDialogFilename');
    let conf = document.getElementById('confirmDeleteName');
    conf.innerHTML = filename.value;
    showPopUpById('deleteConfirm')
}

function deleteFile() {
    let filename = document.getElementById('saveDialogFilename');
    let formData = new FormData();
    formData.append('cmd', "delete");
    formData.append('name', filename.value);
    fetchHelperForm("/files/", formData, text => {
        if (text.trim() !== "true") {
            showPopUpById('deleteError')
            return;
        }
        hidePopUp();
    })
}

function fetchHelper(url, data, target) {
    let formData = new FormData();
    formData.append('data', data);
    fetchHelperForm(url, formData, target);
}

function fetchHelperForm(url, formData, target) {
    fetch(url, {body: formData, method: "post", signal: AbortSignal.timeout(10000)})
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
