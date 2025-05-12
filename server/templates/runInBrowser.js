function runSource() {
    let source = document.getElementById('source');
    let result = document.getElementById('result');

    result.innerHTML = generateOutput(source.value);
    source.focus();
}

let myWindow;

function runSourceInWindow() {
    let source = document.getElementById('source');

    if (!(myWindow && !myWindow.closed)) {
        myWindow = window.open("", "", "width=810,height=620");
    }

    myWindow.document.body.innerHTML = generateOutput(source.value);
    source.focus();
}
