function updateByGui(n) {
    let slValues = ""
    for (let i = 0; i < n; i++) {
        let sl = document.getElementById('guiElement-' + i);
        if (slValues !== "") {
            slValues += ",";
        }
        if (sl.type === "checkbox") {
            slValues += sl.checked ? "true" : "false";
        } else {
            slValues += sl.value;
        }
    }

    let source = document.getElementById('gui-source');
    let result = document.getElementById('gui-inner');

    result.innerHTML = generateOutput(source.value, slValues);
}

function runSource() {
    let source = document.getElementById('source');
    let result = document.getElementById('result');

    result.innerHTML = generateOutput(source.value);
    result.scrollTop = result.scrollHeight;
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
        "  <script src=\"/assets/wasm_exec.js\"></script>\n" +
        "  <script>\n" +
        "    const go = new Go();\n" +
        "\tWebAssembly.instantiateStreaming(fetch(\"/assets/generate.wasm\"), go.importObject).then((result) => {\n" +
        "\t  go.run(result.instance);\n" +
        "\t});\n" +
        "  </script>\n"+
        "  <script type=\"text/javascript\" src=\"/assets/runInBrowser.js\"></script>\n"+
        "  <script type=\"text/javascript\" src=\"/assets/main.js\"></script>\n"+
        "</head>\n" +
        "<body>\n"+
        "<div class=\"div-main\">\n");
    myWindow.document.write(generateOutput(source.value));
    myWindow.document.write("\n</div></body>\n</html>");
    source.focus();
}
