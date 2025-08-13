function slider(n) {
    let slValues = ""
    for (let i = 0; i < n; i++) {
        let sl = document.getElementById('slider-' + i);
        if (slValues !== "") {
            slValues += ",";
        }
        slValues += sl.value;
    }

    let source = document.getElementById('source');
    let result = document.getElementById('slider-inner');

    result.innerHTML = generateOutput(source.value, slValues);
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
        "  <script src=\"/assets/wasm_exec.js\"></script>\n" +
        "  <script>\n" +
        "    const go = new Go();\n" +
        "\tWebAssembly.instantiateStreaming(fetch(\"/assets/generate.wasm\"), go.importObject).then((result) => {\n" +
        "\t  go.run(result.instance);\n" +
        "\t});\n" +
        "  </script>\n"+
        "  <script type=\"text/javascript\" src=\"/js/execute.js\"></script>\n"+
        "  <script type=\"text/javascript\" src=\"/assets/main.js\"></script>\n"+
        "</head>\n" +
        "<body>\n"+
        "<textarea id=\"source\" style=\"display:none;\">" + source.value + "</textarea>\n");
    myWindow.document.write(generateOutput(source.value));
    myWindow.document.write("\n</body>\n</html>");
    source.focus();
}
