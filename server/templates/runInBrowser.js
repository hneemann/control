function slider(n) {
    let slValues=""
    for (let i = 0; i < n; i++) {
        let sl = document.getElementById('slider-' + i);
        if (slValues !== "") {
            slValues += ",";
        }
        slValues += sl.value;
    }

    let source = document.getElementById('source');
    let result = document.getElementById('result');

    result.innerHTML = generateOutput(source.value, slValues);
    source.focus();
}

function runSource() {
    let source = document.getElementById('source');
    let result = document.getElementById('result');

    result.innerHTML = generateOutput(source.value);
    source.focus();
}

let myWindow;

function runSourceInWindow() {
    let source = document.getElementById('source');

    if (myWindow && !myWindow.closed) {
        myWindow.document.body.innerHTML = "";
    } else {
        myWindow = window.open("", "", "width=810,height=620");
    }

    myWindow.document.write("<head>\n" +
        "  <meta charset=\"UTF-8\">\n" +
        "  <title>Control</title>\n" +
        "  <link rel=\"icon\" type=\"image/svg\" href=\"/assets/icon.svg\">\n" +
        "  <link rel=\"stylesheet\" type=\"text/css\" href=\"/assets/window.css\"/>\n" +
        "</head>\n" +
        "<body>\n");
    myWindow.document.write(generateOutput(source.value));
    myWindow.document.write("\n</body>\n</html>");
    source.focus();
}
